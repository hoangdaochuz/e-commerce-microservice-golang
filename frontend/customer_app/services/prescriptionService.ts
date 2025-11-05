import { api } from '@/lib/axios';

// Types
export interface Prescription {
  id: string;
  patientId: string;
  doctorId: string;
  doctorName: string;
  medication: string;
  dosage: string;
  frequency: string;
  duration: string;
  instructions: string;
  startDate: string;
  endDate: string;
  status: 'active' | 'completed' | 'cancelled';
  refillsRemaining: number;
  createdAt: string;
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

// Prescription Service
export const prescriptionService = {
  /**
   * Get all prescriptions for a patient
   */
  getPrescriptions: async (
    patientId: string,
    params?: {
      page?: number;
      limit?: number;
      status?: string;
    }
  ): Promise<PaginatedResponse<Prescription>> => {
    const response = await api.get<ApiResponse<PaginatedResponse<Prescription>>>(
      `/patients/${patientId}/prescriptions`,
      { params }
    );
    return response.data.data;
  },

  /**
   * Get prescription by ID
   */
  getPrescription: async (prescriptionId: string): Promise<Prescription> => {
    const response = await api.get<ApiResponse<Prescription>>(
      `/prescriptions/${prescriptionId}`
    );
    return response.data.data;
  },

  /**
   * Request prescription refill
   */
  requestRefill: async (prescriptionId: string): Promise<Prescription> => {
    const response = await api.post<ApiResponse<Prescription>>(
      `/prescriptions/${prescriptionId}/refill`
    );
    return response.data.data;
  },

  /**
   * Get active prescriptions
   */
  getActivePrescriptions: async (patientId: string): Promise<Prescription[]> => {
    const response = await api.get<ApiResponse<Prescription[]>>(
      `/patients/${patientId}/prescriptions/active`
    );
    return response.data.data;
  },

  /**
   * Download prescription PDF
   */
  downloadPrescription: async (prescriptionId: string): Promise<Blob> => {
    const response = await api.get(`/prescriptions/${prescriptionId}/download`, {
      responseType: 'blob',
    });
    return response.data;
  },
};

