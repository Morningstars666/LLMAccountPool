import React, { useEffect, useState } from 'react';
import { Card, Table, Button, Space, Modal, Form, Input, Select, Upload, message } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, UploadOutlined, CopyOutlined } from '@ant-design/icons';
import { modelApi } from '../services/api';
import { copyToClipboard, downloadBlob } from '../utils';
import type { Model } from '../types';

const Models: React.FC = () => {
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalVisible, setModalVisible] = useState(false);
  const [importModalVisible, setImportModalVisible] = useState(false);
  const [editingModel, setEditingModel] = useState<Model | null>(null);
  const [form] = Form.useForm();
  const [importFile, setImportFile] = useState<File | null>(null);

  const loadModels = async () => {
    try {
      const data = await modelApi.getAll();
      setModels(data);
    } catch (error) {
      console.error('Failed to load models:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadModels();
  }, []);

  const handleAdd = () => {
    setEditingModel(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (record: Model) => {
    setEditingModel(record);
    form.setFieldsValue({
      name: record.name,
      model: record.model,
      strategy: record.strategy,
    });
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm('确定要删除这个模型吗？')) return;
    try {
      await modelApi.delete(id);
      message.success('模型删除成功');
      loadModels();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '删除失败');
    }
  };

  const handleSubmit = async (values: { name: string; model: string; strategy: string }) => {
    try {
      if (editingModel) {
        await modelApi.update(editingModel.id, values);
        message.success('模型更新成功');
      } else {
        await modelApi.create(values);
        message.success('模型创建成功');
      }
      setModalVisible(false);
      loadModels();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '操作失败');
    }
  };

  const handleImport = async () => {
    if (!importFile) {
      message.error('请选择文件');
      return;
    }
    try {
      const result = await modelApi.importExcel(importFile);
      message.success(`导入成功：新建${result.external_models_created}个，更新${result.external_models_updated}个`);
      setImportModalVisible(false);
      setImportFile(null);
      loadModels();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '导入失败');
    }
  };

  const handleDownloadTemplate = async () => {
    try {
      const blob = await modelApi.downloadTemplate();
      downloadBlob(blob, 'models_template.xlsx');
    } catch (error) {
      message.error('下载模板失败');
    }
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => (
        <Space>
          <span>{name}</span>
          <Button type="text" icon={<CopyOutlined />} size="small" onClick={() => copyToClipboard(name)} />
        </Space>
      ),
    },
    {
      title: '模型标识',
      dataIndex: 'model',
      key: 'model',
      render: (model: string) => (
        <Space>
          <span style={{ fontFamily: 'monospace' }}>{model}</span>
          <Button type="text" icon={<CopyOutlined />} size="small" onClick={() => copyToClipboard(model)} />
        </Space>
      ),
    },
    {
      title: '策略',
      dataIndex: 'strategy',
      key: 'strategy',
      render: (strategy: string) => strategy === 'round_robin' ? '轮询切换' : '用完切换',
    },
    {
      title: '请求源数',
      dataIndex: 'sources',
      key: 'source_count',
      render: (sources: Model['sources']) => sources?.length || 0,
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: Model) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>编辑</Button>
          <Button size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="page-header">
        <h2>对外模型</h2>
        <p>管理供客户端调用的模型配置</p>
      </div>

      <Card>
        <div style={{ marginBottom: 16 }}>
          <Button icon={<UploadOutlined />} onClick={() => setImportModalVisible(true)} style={{ marginRight: 8 }}>
            导入xlsx
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            添加模型
          </Button>
        </div>
        <Table columns={columns} dataSource={models.map(m => ({ ...m, key: m.id }))} loading={loading} pagination={false} />
      </Card>

      <Modal
        title={editingModel ? '编辑模型' : '添加模型'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="如：GPT-4" />
          </Form.Item>
          <Form.Item name="model" label="模型标识" rules={[{ required: true }]}>
            <Input placeholder="如：gpt-4" />
          </Form.Item>
          <Form.Item name="strategy" label="负载均衡策略" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="round_robin">轮询切换</Select.Option>
              <Select.Option value="sequential">用完切换</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="导入模型"
        open={importModalVisible}
        onCancel={() => setImportModalVisible(false)}
        onOk={handleImport}
      >
        <p style={{ marginBottom: 15, color: '#666' }}>
          请上传 xlsx 文件，支持导入对外模型和上游模型。
          <a onClick={handleDownloadTemplate} style={{ marginLeft: 8 }}>下载模板文件</a>
        </p>
        <Upload beforeUpload={(file) => { setImportFile(file); return false; }}>
          <Button icon={<UploadOutlined />}>选择文件</Button>
        </Upload>
        {importFile && <div style={{ marginTop: 8 }}>{importFile.name}</div>}
      </Modal>
    </div>
  );
};

export default Models;