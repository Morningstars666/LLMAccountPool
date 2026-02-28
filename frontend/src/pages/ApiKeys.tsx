import React, { useEffect, useState } from 'react';
import { Card, Table, Button, Space, Modal, Form, Input, Select, message } from 'antd';
import { PlusOutlined, DeleteOutlined, ReloadOutlined, EyeOutlined, EyeInvisibleOutlined, CopyOutlined } from '@ant-design/icons';
import { keyApi, modelApi } from '../services/api';
import { copyToClipboard, maskKey } from '../utils';
import type { ApiKey, Model } from '../types';

const ApiKeys: React.FC = () => {
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalVisible, setModalVisible] = useState(false);
  const [displayModalVisible, setDisplayModalVisible] = useState(false);
  const [newKey, setNewKey] = useState('');
  const [visibleKeys, setVisibleKeys] = useState<Set<number>>(new Set());
  const [form] = Form.useForm();

  const loadData = async () => {
    try {
      const [keysData, modelData] = await Promise.all([
        keyApi.getAll(),
        modelApi.getAll(),
      ]);
      setApiKeys(keysData);
      setModels(modelData);
    } catch (error) {
      console.error('Failed to load keys:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleAdd = () => {
    form.resetFields();
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm('确定要删除这个 API Key 吗？')) return;
    try {
      await keyApi.delete(id);
      message.success('API Key 删除成功');
      loadData();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '删除失败');
    }
  };

  const handleReset = async (id: number) => {
    if (!window.confirm('确定要重置这个 API Key 的用量吗？')) return;
    try {
      await keyApi.reset(id);
      message.success('API Key 用量已重置');
      loadData();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '重置失败');
    }
  };

  const handleSubmit = async (values: { external_model_id?: number; note?: string }) => {
    try {
      const result = await keyApi.create({
        external_model_id: values.external_model_id || null,
        note: values.note || '',
      });
      setNewKey(result.key);
      setModalVisible(false);
      setDisplayModalVisible(true);
      loadData();
      message.success('API Key 生成成功');
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '生成失败');
    }
  };

  const toggleKeyVisibility = (id: number) => {
    const newVisible = new Set(visibleKeys);
    if (newVisible.has(id)) {
      newVisible.delete(id);
    } else {
      newVisible.add(id);
    }
    setVisibleKeys(newVisible);
  };

  const getModelName = (externalModelId: number | null) => {
    if (externalModelId === 0 || externalModelId === null) return '全部模型';
    const model = models.find(m => m.id === externalModelId);
    return model?.name || '-';
  };

  const columns = [
    {
      title: 'Key',
      dataIndex: 'key',
      key: 'key',
      render: (key: string, record: ApiKey) => (
        <Space>
          <code style={{ maxWidth: 280, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', display: 'inline-block' }}>
            {visibleKeys.has(record.id) ? key : maskKey(key)}
          </code>
          <Button
            type="text"
            icon={visibleKeys.has(record.id) ? <EyeInvisibleOutlined /> : <EyeOutlined />}
            size="small"
            onClick={() => toggleKeyVisibility(record.id)}
          />
          <Button type="text" icon={<CopyOutlined />} size="small" onClick={() => copyToClipboard(key)} />
        </Space>
      ),
    },
    {
      title: '备注',
      dataIndex: 'note',
      key: 'note',
      render: (note: string) => note || '-',
    },
    {
      title: '对应模型',
      dataIndex: 'external_model_id',
      key: 'external_model_id',
      render: (id: number | null) => getModelName(id),
    },
    {
      title: '调用次数',
      dataIndex: 'used_count',
      key: 'used_count',
    },
    {
      title: 'Input Tokens',
      dataIndex: 'input_tokens',
      key: 'input_tokens',
    },
    {
      title: 'Output Tokens',
      dataIndex: 'output_tokens',
      key: 'output_tokens',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: ApiKey) => (
        <Space>
          <Button size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record.id)}>删除</Button>
          <Button size="small" icon={<ReloadOutlined />} onClick={() => handleReset(record.id)}>重置</Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="page-header">
        <h2>API Key 管理</h2>
        <p>生成和管理客户端调用密钥</p>
      </div>

      <Card>
        <div style={{ marginBottom: 16 }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            生成 Key
          </Button>
        </div>
        <Table columns={columns} dataSource={apiKeys.map(k => ({ ...k, rowKey: k.id }))} loading={loading} pagination={false} rowKey="rowKey" />
      </Card>

      <Modal
        title="生成 API Key"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="external_model_id" label="对应模型">
            <Select placeholder="全部模型" allowClear>
              {models.map(m => <Select.Option key={m.id} value={m.id}>{m.name}</Select.Option>)}
            </Select>
          </Form.Item>
          <Form.Item name="note" label="备注">
            <Input placeholder="可选备注" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="API Key 已生成"
        open={displayModalVisible}
        onCancel={() => setDisplayModalVisible(false)}
        footer={[
          <Button key="copy" icon={<CopyOutlined />} onClick={() => copyToClipboard(newKey)}>
            复制
          </Button>,
          <Button key="close" type="primary" onClick={() => setDisplayModalVisible(false)}>
            关闭
          </Button>,
        ]}
      >
        <div style={{ background: '#fafbfc', padding: '1.25rem', borderRadius: 12, border: '1px solid #e5e7eb', wordBreak: 'break-all', fontFamily: 'monospace', fontSize: '0.8125rem', color: '#059669' }}>
          {newKey}
        </div>
      </Modal>
    </div>
  );
};

export default ApiKeys;