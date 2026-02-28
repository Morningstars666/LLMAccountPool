import React, { useEffect, useState } from 'react';
import { Card, Table, Button, Space, Modal, Form, Input, Select, message } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, EditOutlined as RenameOutlined } from '@ant-design/icons';
import { sourceApi, modelApi } from '../services/api';
import { getLimitText, getStatus } from '../utils';
import type { Source, Model } from '../types';

const Sources: React.FC = () => {
  const [sources, setSources] = useState<Source[]>([]);
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalVisible, setModalVisible] = useState(false);
  const [nameModalVisible, setNameModalVisible] = useState(false);
  const [editingSource, setEditingSource] = useState<Source | null>(null);
  const [editingNameId, setEditingNameId] = useState<number | null>(null);
  const [form] = Form.useForm();
  const [nameForm] = Form.useForm();

  const loadData = async () => {
    try {
      const [sourceData, modelData] = await Promise.all([
        sourceApi.getAll(),
        modelApi.getAll(),
      ]);
      setSources(sourceData);
      setModels(modelData);
    } catch (error) {
      console.error('Failed to load sources:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleAdd = () => {
    setEditingSource(null);
    form.resetFields();
    form.setFieldsValue({ billing_mode: 'count', limit: 0 });
    setModalVisible(true);
  };

  const handleEdit = (record: Source) => {
    setEditingSource(record);
    form.setFieldsValue({
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
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await sourceApi.delete(id);
      message.success('请求源删除成功');
      loadData();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '删除失败');
    }
  };

  const handleReset = async (id: number) => {
    try {
      await sourceApi.reset(id);
      message.success('请求源用量已重置');
      loadData();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '重置失败');
    }
  };

  const handleRename = (record: Source) => {
    setEditingNameId(record.id);
    nameForm.setFieldsValue({ name: record.name });
    setNameModalVisible(true);
  };

  const handleSubmit = async (values: {
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
      setModalVisible(false);
      loadData();
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
      loadData();
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
      render: (name: string, record: Source) => (
        <Space>
          <span>{name}</span>
          <Button type="text" icon={<RenameOutlined />} size="small" onClick={() => handleRename(record)} />
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
          <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>编辑</Button>
          <Button size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record.id)}>删除</Button>
          <Button size="small" icon={<ReloadOutlined />} onClick={() => handleReset(record.id)}>重置</Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="page-header">
        <h2>请求源管理</h2>
        <p>管理上游 LLM API 请求源</p>
      </div>

      <Card>
        <div style={{ marginBottom: 16 }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            添加请求源
          </Button>
        </div>
        <Table columns={columns} dataSource={sources.map(s => ({ ...s, key: s.id }))} loading={loading} pagination={false} scroll={{ x: 'max-content' }} tableLayout="fixed" />
      </Card>

      <Modal
        title={editingSource ? '编辑请求源' : '添加请求源'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="external_model_id" label="所属模型" rules={[{ required: true }]}>
            <Select placeholder="请选择模型">
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
          <Space style={{ width: '100%' }}>
            <Form.Item name="billing_mode" label="计费模式" style={{ width: '50%' }}>
              <Select style={{ width: '100%' }}>
                <Select.Option value="count">按次</Select.Option>
                <Select.Option value="token">按Token</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item name="limit" label="限额" style={{ width: '50%' }}>
              <Input type="number" placeholder="0 表示不限制" />
            </Form.Item>
          </Space>
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

export default Sources;