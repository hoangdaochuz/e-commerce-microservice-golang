import { api } from '@/lib/axios';
import { Patient } from '@/types/patient';

// API Response types
interface UpdatePatientRequest {
  firstName?: string;
  lastName?: string;
  email?: string;
  phone?: string;
  dateOfBirth?: string;
  avatar?: string;
  address?: {
    street?: string;
    city?: string;
    state?: string;
    zipCode?: string;
  };
}

interface ApiResponse<T> {
  data: T;
  message?: string;
}

// Patient Service
export const patientService = {
  /**
   * Get patient profile by ID
   */
  getPatient: async (patientId: string): Promise<Patient> => {
    const response = await api.get<ApiResponse<Patient>>(`/patients/${patientId}`);
    return response.data.data;
  },

  /**
   * Update patient profile
   */
  updatePatient: async (patientId: string, data: UpdatePatientRequest): Promise<Patient> => {
    const response = await api.put<ApiResponse<Patient>>(`/patients/${patientId}`, data);
    return response.data.data;
  },

  /**
   * Upload patient avatar
   */
  uploadAvatar: async (patientId: string, file: File): Promise<{ avatarUrl: string }> => {
    const formData = new FormData();
    formData.append('avatar', file);

    const response = await api.post<ApiResponse<{ avatarUrl: string }>>(
      `/patients/${patientId}/avatar`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    );

    return response.data.data;
  },

  /**
   * Get patient medical history
   */
  getMedicalHistory: async (patientId: string): Promise<string[]> => {
    const response = await api.get<ApiResponse<string[]>>(`/patients/${patientId}/medical-history`);
    return response.data.data;
  },

  /**
   * Add medical history entry
   */
  addMedicalHistory: async (patientId: string, condition: string): Promise<string[]> => {
    const response = await api.post<ApiResponse<string[]>>(
      `/patients/${patientId}/medical-history`,
      { condition }
    );
    return response.data.data;
  },

  /**
   * Get patient allergies
   */
  getAllergies: async (patientId: string): Promise<string[]> => {
    const response = await api.get<ApiResponse<string[]>>(`/patients/${patientId}/allergies`);
    return response.data.data;
  },

  /**
   * Add allergy
   */
  addAllergy: async (patientId: string, allergy: string): Promise<string[]> => {
    const response = await api.post<ApiResponse<string[]>>(
      `/patients/${patientId}/allergies`,
      { allergy }
    );
    return response.data.data;
  },

  /**
   * Get patient medications
   */
  getMedications: async (patientId: string): Promise<string[]> => {
    const response = await api.get<ApiResponse<string[]>>(`/patients/${patientId}/medications`);
    return response.data.data;
  },

  /**
   * Add medication
   */
  addMedication: async (patientId: string, medication: string): Promise<string[]> => {
    const response = await api.post<ApiResponse<string[]>>(
      `/patients/${patientId}/medications`,
      { medication }
    );
    return response.data.data;
  },

  /**
   * Delete patient account
   */
  deletePatient: async (patientId: string): Promise<{ message: string }> => {
    const response = await api.delete<ApiResponse<{ message: string }>>(`/patients/${patientId}`);
    return response.data.data;
  },
};

