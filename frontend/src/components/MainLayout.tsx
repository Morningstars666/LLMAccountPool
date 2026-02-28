import React from 'react';
import { Layout, Menu, Button } from 'antd';
import { 
  DashboardOutlined, 
  ApiOutlined, 
  CloudServerOutlined, 
  KeyOutlined, 
  BarChartOutlined, 
  SettingOutlined,
  LogoutOutlined 
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import './MainLayout.css';

const { Sider, Content } = Layout;

const menuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '仪表盘' },
  { key: '/models', icon: <ApiOutlined />, label: '对外模型' },
  { key: '/sources', icon: <CloudServerOutlined />, label: '请求源管理' },
  { key: '/keys', icon: <KeyOutlined />, label: 'API Key' },
  { key: '/usage', icon: <BarChartOutlined />, label: '用量统计' },
  { key: '/settings', icon: <SettingOutlined />, label: '设置' },
];

const MainLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { logout } = useAuth();

  const handleMenuClick = (key: string) => {
    navigate(key);
  };

  return (
    <Layout className="main-layout">
      <Sider width={260} className="sidebar">
        <div className="sidebar-logo">
          <h2>LLM 账号池</h2>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => handleMenuClick(key)}
          className="sidebar-menu"
        />
        <div className="sidebar-footer">
          <Button 
            type="text" 
            icon={<LogoutOutlined />} 
            onClick={logout}
            className="logout-btn"
          >
            退出登录
          </Button>
        </div>
      </Sider>
      <Content className="main-content">
        <Outlet />
      </Content>
    </Layout>
  );
};

export default MainLayout;