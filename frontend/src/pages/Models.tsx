import React, { useEffect, useState } from 'react';
import { Card, Table, Button, Space, Modal, Form, Input, Select, Upload, message } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, UploadOutlined, CopyOutlined, ReloadOutlined, EditOutlined as RenameOutlined, EyeOutlined } from '@ant-design/icons';
import { modelApi, sourceApi } from '../services/api';
import { copyToClipboard, downloadBlob, getLimitText, getStatus } from '../utils';
import type { Model, Source } from '../types';

const Models: React.FC = () => {
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalVisible, setModalVisible] = useState(false);
  const [importModalVisible, setImportModalVisible] = useState(false);
  const [editingModel, setEditingModel] = useState<Model | null>(null);
  const [form] = Form.useForm();
  const [importFile, setImportFile] = useState<File | null>(null);

  const [sourcesModalVisible, setSourcesModalVisible] = useState(false);
  const [sourceModalVisible, setSourceModalVisible] = useState(false);
  const [nameModalVisible, setNameModalVisible] = useState(false);
  const [selectedModel, setSelectedModel] = useState<Model | null>(null);
  const [sources, setSources] = useState<Source[]>([]);
  const [sourcesLoading, setSourcesLoading] = useState(false);
  const [editingSource, setEditingSource] = useState<Source | null>(null);
  const [editingNameId, setEditingNameId] = useState<number | null>(null);
  const [sourceForm] = Form.useForm();
  const [nameForm] = Form.useForm();

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

  const loadSources = async (modelId: number) => {
    try {
      setSourcesLoading(true);
      const data = await sourceApi.getAll();
      setSources(data.filter(s => s.external_model_id === modelId));
    } catch (error) {
      console.error('Failed to load sources:', error);
    } finally {
      setSourcesLoading(false);
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

  const handleViewSources = (record: Model) => {
    setSelectedModel(record);
    setSourcesModalVisible(true);
    loadSources(record.id);
  };

  const handleSourceAdd = () => {
    setEditingSource(null);
    sourceForm.resetFields();
    sourceForm.setFieldsValue({ 
      external_model_id: selectedModel?.id,
      billing_mode: 'count', 
      limit: 0 
    });
    setSourceModalVisible(true);
  };

  const handleSourceEdit = (record: Source) => {
    setEditingSource(record);
    sourceForm.setFieldsValue({
      external_model_id: record.external_model_id,
      name: record.name,
      api_url: record.api_url,
      api_key: record.api_key,
      model_name: record.model_name,
      billing_mode: record.billing_mode,
      limit: record.billing_mode === 'count' ? record.limit_count : record.limit_tokens,
      limit_reset_interval: record.limit_reset_interval,
      limit_reset_time: record.limit_reset_time,
    });
    setSourceModalVisible(true);
  };

  const handleSourceDelete = async (id: number) => {
    try {
      await sourceApi.delete(id);
      message.success('请求源删除成功');
      selectedModel && loadSources(selectedModel.id);
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '删除失败');
    }
  };

  const handleSourceReset = async (id: number) => {
    try {
      await sourceApi.reset(id);
      message.success('请求源用量已重置');
      selectedModel && loadSources(selectedModel.id);
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '重置失败');
    }
  };

  const handleSourceRename = (record: Source) => {
    setEditingNameId(record.id);
    nameForm.setFieldsValue({ name: record.name });
    setNameModalVisible(true);
  };

  const handleSourceSubmit = async (values: {
    external_model_id: number;
    name: string;
    api_url: string;
    api_key: string;
    model_name: string;
    billing_mode: 'count' | 'token';
    limit: number;
    limit_reset_interval?: number;
    limit_reset_time?: string;
  }) => {
    try {
      const data = {
        external_model_id: values.external_model_id,
        name: values.name,
        api_url: values.api_url,
        api_key: values.api_key,
        model_name: values.model_name,
        billing_mode: values.billing_mode as 'count' | 'token',
        limit_count: values.billing_mode === 'count' ? values.limit : 0,
        limit_tokens: values.billing_mode === 'token' ? values.limit : 0,
        limit_reset_interval: values.limit_reset_interval || 0,
        limit_reset_time: values.limit_reset_time || '',
      };
      
      if (editingSource) {
        await sourceApi.update(editingSource.id, data);
        message.success('请求源更新成功');
      } else {
        await sourceApi.create(data);
        message.success('请求源创建成功');
      }
      setSourceModalVisible(false);
      selectedModel && loadSources(selectedModel.id);
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '操作失败');
    }
  };

  const handleNameSubmit = async (values: { name: string }) => {
    try {
      await sourceApi.updateName(editingNameId!, values.name);
      message.success('请求源名称修改成功');
      setNameModalVisible(false);
      selectedModel && loadSources(selectedModel.id);
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '修改失败');
    }
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
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
      width: 150,
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
      width: 100,
      render: (strategy: string) => strategy === 'round_robin' ? '轮询切换' : '用完切换',
    },
    {
      title: '请求源数',
      dataIndex: 'sources',
      key: 'source_count',
      width: 100,
      render: (sources: Model['sources']) => sources?.length || 0,
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: unknown, record: Model) => (
        <Space size="small">
          <Button size="small" icon={<EyeOutlined />} onClick={() => handleViewSources(record)}>请求源</Button>
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
        <Table columns={columns} dataSource={models.map(m => ({ ...m, key: m.id }))} loading={loading} pagination={false} scroll={{ x: 'max-content' }} tableLayout="fixed" />
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

      <Modal
        title={`请求源 - ${selectedModel?.name || ''}`}
        open={sourcesModalVisible}
        onCancel={() => { setSourcesModalVisible(false); setSelectedModel(null); }}
        footer={null}
        width={1000}
        styles={{ body: { maxHeight: '70vh', overflow: 'auto' } }}
      >
        <div style={{ marginBottom: 16 }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleSourceAdd}>
            添加请求源
          </Button>
        </div>
        <Table
          columns={[
            {
              title: '名称',
              dataIndex: 'name',
              key: 'name',
              width: 150,
              render: (name: string, record: Source) => (
                <Space>
                  <span>{name}</span>
                  <Button type="text" icon={<RenameOutlined />} size="small" onClick={() => handleSourceRename(record)} />
                </Space>
              ),
            },
            {
              title: 'API 地址',
              dataIndex: 'api_url',
              key: 'api_url',
              width: 200,
              ellipsis: true,
              render: (url: string) => <code>{url}</code>,
            },
            {
              title: '模型名称',
              dataIndex: 'model_name',
              key: 'model_name',
              width: 120,
              ellipsis: true,
              render: (name: string) => <code>{name}</code>,
            },
            {
              title: '计费模式',
              dataIndex: 'billing_mode',
              key: 'billing_mode',
              width: 100,
              render: (mode: string) => mode === 'count' ? '按次计费' : '按Token计费',
            },
            {
              title: '已用/限额',
              key: 'usage',
              width: 120,
              render: (_: unknown, record: Source) => getLimitText(record),
            },
            {
              title: '上次重置',
              dataIndex: 'last_reset_at',
              key: 'last_reset_at',
              width: 160,
              render: (text: string) => text || '-',
            },
            {
              title: '状态',
              key: 'status',
              width: 80,
              render: (_: unknown, record: Source) => {
                const { status, color } = getStatus(record);
                return <span style={{ color: color === 'success' ? '#52c41a' : color === 'warning' ? '#faad14' : '#ff4d4f' }}>{status}</span>;
              },
            },
            {
              title: '操作',
              key: 'action',
              width: 180,
              render: (_: unknown, record: Source) => (
                <Space size="small">
                  <Button size="small" icon={<EditOutlined />} onClick={() => handleSourceEdit(record)}>编辑</Button>
                  <Button size="small" danger icon={<DeleteOutlined />} onClick={() => handleSourceDelete(record.id)}>删除</Button>
                  <Button size="small" icon={<ReloadOutlined />} onClick={() => handleSourceReset(record.id)}>重置</Button>
                </Space>
              ),
            },
          ]}
          dataSource={sources.map(s => ({ ...s, key: s.id }))}
          loading={sourcesLoading}
          pagination={false}
          scroll={{ x: 'max-content' }}
          tableLayout="fixed"
        />
      </Modal>

      <Modal
        title={editingSource ? '编辑请求源' : '添加请求源'}
        open={sourceModalVisible}
        onCancel={() => setSourceModalVisible(false)}
        onOk={() => sourceForm.submit()}
        width={600}
      >
        <Form form={sourceForm} layout="vertical" onFinish={handleSourceSubmit}>
          <Form.Item name="external_model_id" label="所属模型" rules={[{ required: true }]}>
            <Select placeholder="请选择模型" disabled>
              {models.map(m => <Select.Option key={m.id} value={m.id}>{m.name}</Select.Option>)}
            </Select>
          </Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="如：OpenAI API" />
          </Form.Item>
          <Form.Item name="api_url" label="API 地址" rules={[{ required: true }]}>
            <Input placeholder="如：https://api.openai.com/v1" />
          </Form.Item>
          <Form.Item name="api_key" label="API Key" rules={[{ required: true }]}>
            <Input placeholder="请输入 API Key" />
          </Form.Item>
          <Form.Item name="model_name" label="上游模型名" rules={[{ required: true }]}>
            <Input placeholder="如：gpt-4" />
          </Form.Item>
          <>
            <Form.Item name="billing_mode" label="计费模式" style={{ width: '50%' }}>
              <Select style={{ width: '100%' }}>
                <Select.Option value="count">按次</Select.Option>
                <Select.Option value="token">按Token</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item name="limit" label="限额" style={{ width: '50%' }}>
              <Input type="number" placeholder="0 表示不限制" />
            </Form.Item>
          </>
          <Form.Item name="limit_reset_interval" label="自动重置间隔（秒）">
            <Input type="number" placeholder="0 表示不自动重置，3600 表示每小时重置一次" />
          </Form.Item>
          <Form.Item name="limit_reset_time" label="定时重置时间（HH:MM）">
            <Input placeholder="如 00:00 表示每天凌晨重置" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="修改请求源名称"
        open={nameModalVisible}
        onCancel={() => setNameModalVisible(false)}
        onOk={() => nameForm.submit()}
      >
        <Form form={nameForm} layout="vertical" onFinish={handleNameSubmit}>
          <Form.Item name="name" label="新名称" rules={[{ required: true }]}>
            <Input placeholder="请输入新名称" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Models;