import React, { useState, useEffect } from 'react';
import { message } from 'antd';
import { startTokenRefresh, stopTokenRefresh } from '../services/api';

interface AuthContextType {
  isAuthenticated: boolean;
  login: (token: string) => void;
  logout: () => void;
}

export const AuthContext = React.createContext<AuthContextType>({
  isAuthenticated: false,
  login: () => {},
  logout: () => {},
});

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(() => {
    return !!localStorage.getItem('token');
  });

  useEffect(() => {
    if (isAuthenticated) {
      startTokenRefresh();
    }
    return () => stopTokenRefresh();
  }, [isAuthenticated]);

  const login = (token: string) => {
    localStorage.setItem('token', token);
    setIsAuthenticated(true);
    startTokenRefresh();
  };

  const logout = () => {
    localStorage.removeItem('token');
    setIsAuthenticated(false);
    stopTokenRefresh();
    message.success('已退出登录');
  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => React.useContext(AuthContext);