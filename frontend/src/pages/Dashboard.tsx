import React, { useEffect, useState } from 'react';
import { Card, Table, Button, Typography, Tooltip } from 'antd';
import { CopyOutlined } from '@ant-design/icons';
import { usageApi, serverApi } from '../services/api';
import { copyToClipboard } from '../utils';
import type { UsageStats, ServerInfo } from '../types';
import './Dashboard.css';

const { Text, Title } = Typography;

const Dashboard: React.FC = () => {
  const [serverInfo, setServerInfo] = useState<ServerInfo | null>(null);
  const [usageStats, setUsageStats] = useState<UsageStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadData = async () => {
      try {
        const [server, usage] = await Promise.all([
          serverApi.getInfo(),
          usageApi.get(),
        ]);
        setServerInfo(server);
        setUsageStats(usage);
      } catch (error) {
        console.error('Failed to load dashboard data:', error);
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, []);

  const totalCalls = usageStats?.api_key_stats?.reduce((sum, k) => sum + k.used_count, 0) || 0;

  const modelColumns = [
    { title: '模型名称', dataIndex: 'name', key: 'name' },
    { 
      title: '模型标识', 
      dataIndex: 'model', 
      key: 'model',
      render: (model: string) => <Text code>{model}</Text>
    },
    { 
      title: '策略', 
      dataIndex: 'strategy', 
      key: 'strategy',
      render: (strategy: string) => strategy === 'round_robin' ? '轮询' : '用完切换'
    },
    { 
      title: '请求源数量', 
      dataIndex: 'source_count', 
      key: 'source_count' 
    },
  ];

  const modelData = usageStats?.model_stats?.map((m, idx) => ({
    ...m,
    key: idx,
  })) || [];

  return (
    <div className="dashboard-page">
      <div className="page-header">
        <Title level={2}>仪表盘</Title>
        <Text type="secondary">系统概览与用量统计</Text>
      </div>

      <Card className="proxy-card" title="请求代理地址">
        <div className="proxy-url-box">
          <Text code className="proxy-url">{serverInfo?.proxy_url || '加载中...'}</Text>
          <Tooltip title="复制">
            <Button icon={<CopyOutlined />} onClick={() => serverInfo && copyToClipboard(serverInfo.proxy_url)} />
          </Tooltip>
        </div>
        <Text type="secondary" className="proxy-hint">
          将此地址配置为你的 API 请求地址，在请求时需要在 Header 中添加 Authorization
        </Text>
      </Card>

      <div className="stats-grid">
        <Card className="stat-card">
          <Text type="secondary">对外模型</Text>
          <div className="stat-value">{usageStats?.model_stats?.length || 0}</div>
        </Card>
        <Card className="stat-card">
          <Text type="secondary">请求源</Text>
          <div className="stat-value">{usageStats?.source_stats?.length || 0}</div>
        </Card>
        <Card className="stat-card">
          <Text type="secondary">API Keys</Text>
          <div className="stat-value">{usageStats?.api_key_stats?.length || 0}</div>
        </Card>
        <Card className="stat-card">
          <Text type="secondary">总调用次数</Text>
          <div className="stat-value">{totalCalls}</div>
        </Card>
      </div>

      <Card title="模型用量详情">
        <Table 
          columns={modelColumns} 
          dataSource={modelData} 
          loading={loading}
          pagination={false}
          size="middle"
        />
      </Card>
    </div>
  );
};

export default Dashboard;