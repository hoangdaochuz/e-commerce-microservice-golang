import { api } from '@/lib/axios';

// Types
export interface Appointment {
  id: string;
  patientId: string;
  doctorId: string;
  doctorName: string;
  specialty: string;
  date: string;
  time: string;
  status: 'scheduled' | 'completed' | 'cancelled' | 'no-show';
  reason: string;
  notes?: string;
}

export interface CreateAppointmentRequest {
  doctorId: string;
  date: string;
  time: string;
  reason: string;
  notes?: string;
}

export interface UpdateAppointmentRequest {
  date?: string;
  time?: string;
  reason?: string;
  notes?: string;
  status?: 'scheduled' | 'completed' | 'cancelled' | 'no-show';
}

interface ApiResponse<T> {
  data: T;
  message?: string;
}

interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

// Appointment Service
export const appointmentService = {
  /**
   * Get all appointments for a patient
   */
  getAppointments: async (
    patientId: string,
    params?: {
      page?: number;
      limit?: number;
      status?: string;
      fromDate?: string;
      toDate?: string;
    }
  ): Promise<PaginatedResponse<Appointment>> => {
    const response = await api.get<ApiResponse<PaginatedResponse<Appointment>>>(
      `/patients/${patientId}/appointments`,
      { params }
    );
    return response.data.data;
  },

  /**
   * Get appointment by ID
   */
  getAppointment: async (appointmentId: string): Promise<Appointment> => {
    const response = await api.get<ApiResponse<Appointment>>(`/appointments/${appointmentId}`);
    return response.data.data;
  },

  /**
   * Create new appointment
   */
  createAppointment: async (
    patientId: string,
    data: CreateAppointmentRequest
  ): Promise<Appointment> => {
    const response = await api.post<ApiResponse<Appointment>>(
      `/patients/${patientId}/appointments`,
      data
    );
    return response.data.data;
  },

  /**
   * Update appointment
   */
  updateAppointment: async (
    appointmentId: string,
    data: UpdateAppointmentRequest
  ): Promise<Appointment> => {
    const response = await api.put<ApiResponse<Appointment>>(
      `/appointments/${appointmentId}`,
      data
    );
    return response.data.data;
  },

  /**
   * Cancel appointment
   */
  cancelAppointment: async (appointmentId: string): Promise<Appointment> => {
    const response = await api.patch<ApiResponse<Appointment>>(
      `/appointments/${appointmentId}/cancel`
    );
    return response.data.data;
  },

  /**
   * Delete appointment
   */
  deleteAppointment: async (appointmentId: string): Promise<{ message: string }> => {
    const response = await api.delete<ApiResponse<{ message: string }>>(
      `/appointments/${appointmentId}`
    );
    return response.data.data;
  },

  /**
   * Get available time slots for a doctor
   */
  getAvailableSlots: async (
    doctorId: string,
    date: string
  ): Promise<string[]> => {
    const response = await api.get<ApiResponse<string[]>>(
      `/doctors/${doctorId}/available-slots`,
      { params: { date } }
    );
    return response.data.data;
  },
};

