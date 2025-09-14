import axios from 'axios';

// 创建 axios 实例
export const apiClient = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
apiClient.interceptors.request.use(
  (config: any) => {
    return config;
  },
  (error: any) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
apiClient.interceptors.response.use(
  (response: any) => {
    return response;
  },
  (error: any) => {
    if (error.response) {
      // 服务器响应了错误状态码
      const message = error.response.data?.error || error.response.data?.message || 'Request failed';
      return Promise.reject(new Error(message));
    } else if (error.request) {
      // 请求没有收到响应
      return Promise.reject(new Error('Network error: No response received'));
    } else {
      // 请求配置出错
      return Promise.reject(new Error('Request configuration error'));
    }
  }
);