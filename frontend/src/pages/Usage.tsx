import React, { useEffect, useState } from 'react';
import { Card, Table } from 'antd';
import { usageApi } from '../services/api';
import { getLimitText, getStatus } from '../utils';
import type { UsageStats } from '../types';

const Usage: React.FC = () => {
  const [usageStats, setUsageStats] = useState<UsageStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadData = async () => {
      try {
        const data = await usageApi.get();
        setUsageStats(data);
      } catch (error) {
        console.error('Failed to load usage:', error);
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, []);

  const sourceColumns = [
    { title: '请求源', dataIndex: 'name', key: 'name' },
    { title: '模型', dataIndex: 'model_name', key: 'model_name', render: (name: string) => <code>{name}</code> },
    { title: '计费模式', dataIndex: 'billing_mode', key: 'billing_mode', render: (mode: string) => mode === 'count' ? '按次' : '按Token' },
    { title: '已用次数', dataIndex: 'used_count', key: 'used_count' },
    { title: '已用Token', dataIndex: 'used_tokens', key: 'used_tokens' },
    { title: '限额', key: 'limit', render: (_: unknown, record: UsageStats['source_stats'][0]) => getLimitText(record) },
    { title: '上次重置', dataIndex: 'last_reset_at', key: 'last_reset_at', render: (text: string) => text || '-' },
    {
      title: '状态',
      key: 'status',
      render: (_: unknown, record: UsageStats['source_stats'][0]) => {
        const { status, color } = getStatus(record);
        return <span style={{ color: color === 'success' ? '#52c41a' : color === 'warning' ? '#faad14' : '#ff4d4f' }}>{status}</span>;
      },
    },
  ];

  const keyColumns = [
    { title: 'Key', dataIndex: 'key', key: 'key', render: (key: string) => <code>{key}</code> },
    { title: '备注', dataIndex: 'note', key: 'note', render: (note: string) => note || '-' },
    { title: '调用次数', dataIndex: 'used_count', key: 'used_count' },
    { title: 'Input Tokens', dataIndex: 'input_tokens', key: 'input_tokens' },
    { title: 'Output Tokens', dataIndex: 'output_tokens', key: 'output_tokens' },
    { title: '总Token', dataIndex: 'used_tokens', key: 'used_tokens' },
  ];

  return (
    <div>
      <div className="page-header">
        <h2>用量统计</h2>
        <p>查看详细的调用统计数据</p>
      </div>

      <Card title="请求源用量" style={{ marginBottom: 16 }}>
        <Table
          columns={sourceColumns}
          dataSource={usageStats?.source_stats?.map((s, idx) => ({ ...s, key: idx })) || []}
          loading={loading}
          pagination={false}
        />
      </Card>

      <Card title="API Key 用量">
        <Table
          columns={keyColumns}
          dataSource={usageStats?.api_key_stats?.map((k, idx) => ({ ...k, key: idx })) || []}
          loading={loading}
          pagination={false}
        />
      </Card>
    </div>
  );
};

export default Usage;