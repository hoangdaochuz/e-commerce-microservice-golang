import { authService } from '@/services';
import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';

// Create axios instance with default config
const axiosInstance: AxiosInstance = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1',
  timeout: parseInt(process.env.NEXT_PUBLIC_API_TIMEOUT || '30000'),
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor
axiosInstance.interceptors.request.use(
  (config) => {
    // Get auth token from localStorage
    if (typeof window !== 'undefined') {
      const tokenKey = process.env.NEXT_PUBLIC_AUTH_TOKEN_KEY || 'auth_token';
      const token = localStorage.getItem(tokenKey);

      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }

    // Log request if logging is enabled
    if (process.env.NEXT_PUBLIC_ENABLE_LOGGING === 'true') {
      console.log('🚀 Request:', {
        method: config.method?.toUpperCase(),
        url: config.url,
        baseURL: config.baseURL,
        params: config.params,
        data: config.data,
      });
    }

    return config;
  },
  (error: AxiosError) => {
    console.error('❌ Request Error:', error);
    return Promise.reject(error);
  }
);

// Response interceptor
axiosInstance.interceptors.response.use(
  (response: AxiosResponse) => {
    // Log response if logging is enabled
    if (process.env.NEXT_PUBLIC_ENABLE_LOGGING === 'true') {
      console.log('✅ Response:', {
        status: response.status,
        statusText: response.statusText,
        url: response.config.url,
        data: response.data,
      });
    }

    return response;
  },
  async (error: AxiosError) => {
    // Handle different error scenarios
    if (error.response) {
      // Server responded with error status
      const status = error.response.status;
      const data = error.response.data;

      console.error('❌ Response Error:', {
        status,
        statusText: error.response.statusText,
        url: error.config?.url,
        data,
      });

      // Handle specific status codes
      switch (status) {
        case 401:
          // Unauthorized 
          // call this func to redirect to login page
          await authService.login({ Username: "" })
          break;

        case 403:
          console.error('Access forbidden');
          break;
        case 404:
          console.error('Resource not found');
          break;
        case 500:
          console.error('Internal server error');
          break;
        default:
          console.error(`Error ${status}:`, data);
      }
    } else if (error.request) {
      // Request was made but no response received
      console.error('❌ Network Error: No response received', error.request);
    } else {
      // Something else happened
      console.error('❌ Error:', error.message);
    }

    return Promise.reject(error);
  }
);

// API helper functions with better typing
export const api = {
  get: <T = any>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> => {
    return axiosInstance.get<T>(url, { withCredentials: true, ...(config || {}) });
  },

  post: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> => {
    return axiosInstance.post<T>(url, data, { withCredentials: true, ...(config || {}) });
  },

  put: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> => {
    return axiosInstance.put<T>(url, data, { withCredentials: true, ...(config || {}) });
  },

  patch: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> => {
    return axiosInstance.patch<T>(url, data, { withCredentials: true, ...(config || {}) });
  },

  delete: <T = any>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> => {
    return axiosInstance.delete<T>(url, { withCredentials: true, ...(config || {}) });
  },
};

// Export configured axios instance
export default axiosInstance;

