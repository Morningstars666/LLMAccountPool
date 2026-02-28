import React, { useState } from 'react';
import { Card, Form, Input, Button, message } from 'antd';
import { authApi } from '../services/api';
import { useAuth } from '../hooks/useAuth';

const Settings: React.FC = () => {
  const { logout } = useAuth();
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [usernameLoading, setUsernameLoading] = useState(false);
  const [passwordForm] = Form.useForm();
  const [usernameForm] = Form.useForm();

  const handlePasswordChange = async (values: { old_password: string; new_password: string; confirm_password: string }) => {
    if (values.new_password !== values.confirm_password) {
      message.error('两次输入的密码不一致');
      return;
    }

    const passwordRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[^a-zA-Z0-9]).{8,}$/;
    if (!passwordRegex.test(values.new_password)) {
      message.error('密码必须至少8位，包含大小写字母、数字和特殊字符');
      return;
    }

    setPasswordLoading(true);
    try {
      await authApi.changePassword(values.old_password, values.new_password, values.confirm_password);
      message.success('密码修改成功，请重新登录');
      setTimeout(() => logout(), 1500);
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '修改失败');
    } finally {
      setPasswordLoading(false);
    }
  };

  const handleUsernameChange = async (values: { password: string; new_username: string }) => {
    if (!values.new_username.trim()) {
      message.error('新用户名不能为空');
      return;
    }

    if (values.new_username.trim() === values.password.trim()) {
      message.error('新用户名不能与密码相同');
      return;
    }

    setUsernameLoading(true);
    try {
      await authApi.changeUsername(values.new_username, values.password);
      message.success('用户名修改成功，请重新登录');
      setTimeout(() => logout(), 1500);
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } };
      message.error(err.response?.data?.error || '修改失败');
    } finally {
      setUsernameLoading(false);
    }
  };

  return (
    <div>
      <div className="page-header">
        <h2>设置</h2>
        <p>修改账号密码</p>
      </div>

      <Card title="修改用户名" style={{ maxWidth: 500, marginBottom: 16 }}>
        <Form form={usernameForm} layout="vertical" onFinish={handleUsernameChange}>
          <Form.Item name="password" label="当前密码" rules={[{ required: true }]}>
            <Input.Password placeholder="请输入当前密码" />
          </Form.Item>
          <Form.Item name="new_username" label="新用户名" rules={[{ required: true }]}>
            <Input placeholder="请输入新用户名" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={usernameLoading}>
              修改用户名
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <Card title="修改密码" style={{ maxWidth: 500 }}>
        <Form form={passwordForm} layout="vertical" onFinish={handlePasswordChange}>
          <Form.Item name="old_password" label="当前密码" rules={[{ required: true }]}>
            <Input.Password placeholder="请输入当前密码" />
          </Form.Item>
          <Form.Item name="new_password" label="新密码" rules={[{ required: true }]}>
            <Input.Password placeholder="请输入新密码（至少8位，包含大小写字母、数字和特殊字符）" />
          </Form.Item>
          <Form.Item name="confirm_password" label="确认新密码" rules={[{ required: true }]}>
            <Input.Password placeholder="请再次输入新密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={passwordLoading}>
              修改密码
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default Settings;