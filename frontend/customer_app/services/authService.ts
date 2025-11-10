import { User } from '@/app/types/patient';
import { api } from '@/lib/axios';

// API Response types
export interface LoginRequest {
  Username: string;
}

export interface RedirectResponse {
  IsSuccess: boolean;
  RedirectURL: string;
}

interface RegisterRequest {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
  phone?: string;
  dateOfBirth?: string;
}

// interface RegisterResponse {
//   token: string;
//   RedirectURL: string;
// }

// Auth Service
export const authService = {
  /**
   * Login with email and password
   */
  login: async (request: LoginRequest): Promise<RedirectResponse> => {
    const response = await api.post<RedirectResponse>('/auth/Login', request);
    console.log("🚀 ~ response:", response)
    if (response.status === 200 && response.data.RedirectURL) {
      // Redirect to Zitadel login page
      window.location.href = response.data.RedirectURL;
    }
    return response.data;
  },

  /**
   * Register new patient account
   */
  // register: async (data: RegisterRequest): Promise<RegisterResponse> => {
  //   const response = await api.post<RegisterResponse>('/auth/register', data);

  //   // Store token in localStorage
  //   if (response.data.token) {
  //     const tokenKey = process.env.NEXT_PUBLIC_AUTH_TOKEN_KEY || 'auth_token';
  //     localStorage.setItem(tokenKey, response.data.token);
  //   }

  //   return response.data;
  // },

  /**
   * Logout and clear stored token
   */
  logout: async (): Promise<void> => {
    try {
      const res = await api.post<RedirectResponse>('/auth/Logout', {});
      if (res.status === 200 && res.data.RedirectURL) {
        // Redirect to Zitadel logout page
        window.location.href = res.data.RedirectURL;
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

