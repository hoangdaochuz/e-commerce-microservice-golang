import { User } from '@/app/types/patient';
import { api } from '@/lib/axios';
import { Patient } from '@/types/patient';

// API Response types
export interface LoginRequest {
  Username: string;
}

export interface LoginResponse {
  token: string;
  patient: Patient;
}

interface RegisterRequest {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
  phone?: string;
  dateOfBirth?: string;
}

interface RegisterResponse {
  token: string;
  patient: Patient;
}

// Auth Service
export const authService = {
  /**
   * Login with email and password
   */
  login: async (request: LoginRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/auth/Login', request);
    if (response.status === 200) {
      if (response.request.responseURL) {
        window.location.href = response.request.responseURL;
      }
    }
    return response.data;
  },

  /**
   * Register new patient account
   */
  register: async (data: RegisterRequest): Promise<RegisterResponse> => {
    const response = await api.post<RegisterResponse>('/auth/register', data);

    // Store token in localStorage
    if (response.data.token) {
      const tokenKey = process.env.NEXT_PUBLIC_AUTH_TOKEN_KEY || 'auth_token';
      localStorage.setItem(tokenKey, response.data.token);
    }

    return response.data;
  },

  /**
   * Logout and clear stored token
   */
  logout: async (): Promise<void> => {
    try {
      const res = await api.post('/auth/Logout', {});
      if (res.status === 200) {
        if (res.request.responseURL) {
          window.location.href = res.request.responseURL;
        }
      }
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // // Clear token from localStorage
      // const tokenKey = process.env.NEXT_PUBLIC_AUTH_TOKEN_KEY || 'auth_token';
      // localStorage.removeItem(tokenKey);
      // localStorage.removeItem('patient');
    }
  },

  /**
   * Get current authenticated patient
   */
  getCurrentPatient: async (): Promise<Patient> => {
    const response = await api.get<Patient>('/auth/me');
    return response.data;
  },

  /**
   * Refresh authentication token
   */
  refreshToken: async (): Promise<{ token: string }> => {
    const response = await api.post<{ token: string }>('/auth/refresh');

    if (response.data.token) {
      const tokenKey = process.env.NEXT_PUBLIC_AUTH_TOKEN_KEY || 'auth_token';
      localStorage.setItem(tokenKey, response.data.token);
    }

    return response.data;
  },

  /**
   * Request password reset
   */
  forgotPassword: async (email: string): Promise<{ message: string }> => {
    const response = await api.post<{ message: string }>('/auth/forgot-password', { email });
    return response.data;
  },

  /**
   * Reset password with token
   */
  resetPassword: async (token: string, newPassword: string): Promise<{ message: string }> => {
    const response = await api.post<{ message: string }>('/auth/reset-password', {
      token,
      password: newPassword,
    });
    return response.data;
  },

  getMe: async () => {
    const response = await api.post<User>("/auth/GetMyProfile", {})
    return response.data
  }
};

