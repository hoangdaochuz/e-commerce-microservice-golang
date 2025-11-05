# API Setup with Axios

This document explains the Axios configuration and API service setup for the Healthcare Portal frontend.

## Table of Contents

1. [Environment Variables](#environment-variables)
2. [Axios Configuration](#axios-configuration)
3. [API Services](#api-services)
4. [Usage Examples](#usage-examples)
5. [Error Handling](#error-handling)

## Environment Variables

### Configuration Files

#### `.env.local` (Local Development)
Contains your actual configuration values. This file is gitignored and should never be committed.

```env
# API Configuration
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_API_TIMEOUT=30000

# Authentication
NEXT_PUBLIC_AUTH_TOKEN_KEY=auth_token

# Feature Flags
NEXT_PUBLIC_ENABLE_LOGGING=true
```

#### `.env.example` (Template)
A template file showing all required environment variables. This should be committed to git.

### Environment Variables Explained

| Variable | Description | Default |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_BASE_URL` | Base URL for your backend API | `http://localhost:8080/api/v1` |
| `NEXT_PUBLIC_API_TIMEOUT` | Request timeout in milliseconds | `30000` (30 seconds) |
| `NEXT_PUBLIC_AUTH_TOKEN_KEY` | LocalStorage key for auth token | `auth_token` |
| `NEXT_PUBLIC_ENABLE_LOGGING` | Enable API request/response logging | `true` |

> **Note:** All variables must be prefixed with `NEXT_PUBLIC_` to be accessible in the browser.

## Axios Configuration

### Location: `/lib/axios.ts`

The Axios instance is configured with:

- **Base URL:** From environment variable
- **Timeout:** Configurable request timeout
- **Headers:** Default JSON content type
- **Interceptors:** Request and response interceptors for auth and error handling

### Features

#### 1. Request Interceptor
- Automatically adds JWT token to all requests
- Logs request details when logging is enabled
- Handles authentication headers

#### 2. Response Interceptor
- Logs response details when logging is enabled
- Handles common HTTP error codes:
  - **401 Unauthorized:** Clears auth and redirects to home
  - **403 Forbidden:** Access denied errors
  - **404 Not Found:** Resource not found
  - **500 Server Error:** Internal server errors
- Network error handling

#### 3. Typed API Helpers
```typescript
api.get<T>(url, config)
api.post<T>(url, data, config)
api.put<T>(url, data, config)
api.patch<T>(url, data, config)
api.delete<T>(url, config)
```

## API Services

All API services are organized by domain in the `/services` directory.

### 1. Auth Service (`authService.ts`)

Handles authentication-related API calls.

**Available Methods:**
- `login(credentials)` - User login
- `register(data)` - User registration
- `logout()` - User logout
- `getCurrentPatient()` - Get authenticated patient
- `refreshToken()` - Refresh JWT token
- `forgotPassword(email)` - Request password reset
- `resetPassword(token, password)` - Reset password

### 2. Patient Service (`patientService.ts`)

Handles patient profile and medical data.

**Available Methods:**
- `getPatient(patientId)` - Get patient profile
- `updatePatient(patientId, data)` - Update patient profile
- `uploadAvatar(patientId, file)` - Upload avatar image
- `getMedicalHistory(patientId)` - Get medical history
- `addMedicalHistory(patientId, condition)` - Add medical condition
- `getAllergies(patientId)` - Get allergies list
- `addAllergy(patientId, allergy)` - Add new allergy
- `getMedications(patientId)` - Get medications list
- `addMedication(patientId, medication)` - Add new medication
- `deletePatient(patientId)` - Delete patient account

### 3. Appointment Service (`appointmentService.ts`)

Manages patient appointments.

**Available Methods:**
- `getAppointments(patientId, params)` - Get all appointments
- `getAppointment(appointmentId)` - Get single appointment
- `createAppointment(patientId, data)` - Create new appointment
- `updateAppointment(appointmentId, data)` - Update appointment
- `cancelAppointment(appointmentId)` - Cancel appointment
- `deleteAppointment(appointmentId)` - Delete appointment
- `getAvailableSlots(doctorId, date)` - Get available time slots

### 4. Prescription Service (`prescriptionService.ts`)

Manages prescriptions and medications.

**Available Methods:**
- `getPrescriptions(patientId, params)` - Get all prescriptions
- `getPrescription(prescriptionId)` - Get single prescription
- `requestRefill(prescriptionId)` - Request prescription refill
- `getActivePrescriptions(patientId)` - Get active prescriptions
- `downloadPrescription(prescriptionId)` - Download prescription PDF

## Usage Examples

### Example 1: Login User

```typescript
'use client';

import { useState } from 'react';
import { authService } from '@/services';
import { useAuth } from '@/contexts/AuthContext';

export default function LoginForm() {
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      const response = await authService.login({ email, password });
      login(response.patient);
      // Redirect or show success
    } catch (err: any) {
      setError(err.response?.data?.message || 'Login failed');
    }
  };

  return (
    <form onSubmit={handleLogin}>
      {error && <p className="text-red-500">{error}</p>}
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Email"
      />
      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Password"
      />
      <button type="submit">Login</button>
    </form>
  );
}
```

### Example 2: Fetch Patient Profile

```typescript
'use client';

import { useEffect, useState } from 'react';
import { patientService } from '@/services';
import { Patient } from '@/types/patient';

export default function ProfilePage() {
  const [patient, setPatient] = useState<Patient | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchPatient = async () => {
      try {
        const data = await patientService.getPatient('patient-id-123');
        setPatient(data);
      } catch (err: any) {
        setError(err.response?.data?.message || 'Failed to load profile');
      } finally {
        setLoading(false);
      }
    };

    fetchPatient();
  }, []);

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;
  if (!patient) return <div>No patient data</div>;

  return (
    <div>
      <h1>{patient.firstName} {patient.lastName}</h1>
      <p>{patient.email}</p>
    </div>
  );
}
```

### Example 3: Create Appointment

```typescript
'use client';

import { useState } from 'react';
import { appointmentService } from '@/services';

export default function BookAppointment({ patientId }: { patientId: string }) {
  const [formData, setFormData] = useState({
    doctorId: '',
    date: '',
    time: '',
    reason: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      const appointment = await appointmentService.createAppointment(
        patientId,
        formData
      );
      
      alert(`Appointment created! ID: ${appointment.id}`);
      // Redirect or update UI
    } catch (err: any) {
      alert(err.response?.data?.message || 'Failed to create appointment');
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* Form inputs */}
      <button type="submit">Book Appointment</button>
    </form>
  );
}
```

### Example 4: Update Patient Profile

```typescript
'use client';

import { useState } from 'react';
import { patientService } from '@/services';
import { useAuth } from '@/contexts/AuthContext';

export default function EditProfile() {
  const { patient, updatePatient } = useAuth();
  const [formData, setFormData] = useState({
    firstName: patient?.firstName || '',
    lastName: patient?.lastName || '',
    phone: patient?.phone || '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!patient) return;

    try {
      const updatedPatient = await patientService.updatePatient(
        patient.id,
        formData
      );
      
      updatePatient(updatedPatient);
      alert('Profile updated successfully!');
    } catch (err: any) {
      alert(err.response?.data?.message || 'Failed to update profile');
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* Form inputs */}
      <button type="submit">Save Changes</button>
    </form>
  );
}
```

### Example 5: Upload Avatar

```typescript
'use client';

import { useState } from 'react';
import { patientService } from '@/services';

export default function AvatarUpload({ patientId }: { patientId: string }) {
  const [uploading, setUploading] = useState(false);

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setUploading(true);
    
    try {
      const result = await patientService.uploadAvatar(patientId, file);
      alert('Avatar uploaded successfully!');
      // Update UI with result.avatarUrl
    } catch (err: any) {
      alert(err.response?.data?.message || 'Failed to upload avatar');
    } finally {
      setUploading(false);
    }
  };

  return (
    <div>
      <input
        type="file"
        accept="image/*"
        onChange={handleFileChange}
        disabled={uploading}
      />
      {uploading && <p>Uploading...</p>}
    </div>
  );
}
```

## Error Handling

### Global Error Handling

The Axios interceptor automatically handles common errors:

```typescript
// 401 Unauthorized - Clears session and redirects
if (status === 401) {
  localStorage.removeItem('auth_token');
  localStorage.removeItem('patient');
  window.location.href = '/';
}
```

### Component-Level Error Handling

Always wrap API calls in try-catch blocks:

```typescript
try {
  const result = await authService.login(credentials);
  // Handle success
} catch (error: any) {
  // Handle error
  const message = error.response?.data?.message || 'Something went wrong';
  console.error('Error:', message);
}
```

### Error Response Structure

Expected error response from backend:

```typescript
{
  error: string;        // Error type
  message: string;      // Human-readable message
  statusCode: number;   // HTTP status code
  details?: any;        // Additional error details
}
```

## Best Practices

1. **Always use TypeScript types** for API responses
2. **Handle errors gracefully** with user-friendly messages
3. **Show loading states** during API calls
4. **Use environment variables** for configuration
5. **Don't commit `.env.local`** to version control
6. **Enable logging in development** for debugging
7. **Disable logging in production** for performance
8. **Implement proper loading and error UI** states
9. **Cache responses** when appropriate
10. **Validate data** before sending to API

## Security Considerations

1. **JWT Tokens:** Stored in localStorage (consider httpOnly cookies for production)
2. **HTTPS Only:** Always use HTTPS in production
3. **Token Refresh:** Implement token refresh logic
4. **CSRF Protection:** Add CSRF tokens if needed
5. **Input Validation:** Always validate user input
6. **Rate Limiting:** Implement on backend
7. **Sensitive Data:** Never log sensitive data in production

## Testing

### Manual Testing

1. Check console logs when `NEXT_PUBLIC_ENABLE_LOGGING=true`
2. Use browser DevTools Network tab
3. Test error scenarios (401, 404, 500, network errors)

### Example Test (Jest)

```typescript
import { authService } from '@/services';
import { api } from '@/lib/axios';

jest.mock('@/lib/axios');

describe('authService', () => {
  it('should login successfully', async () => {
    const mockResponse = {
      data: {
        token: 'mock-token',
        patient: { id: '1', firstName: 'John', lastName: 'Doe' }
      }
    };
    
    (api.post as jest.Mock).mockResolvedValue(mockResponse);
    
    const result = await authService.login({
      email: 'test@example.com',
      password: 'password123'
    });
    
    expect(result.token).toBe('mock-token');
    expect(result.patient.firstName).toBe('John');
  });
});
```

## Troubleshooting

### Common Issues

1. **CORS Errors**
   - Ensure backend allows requests from your frontend origin
   - Check backend CORS configuration

2. **401 Unauthorized**
   - Check if token is being sent in headers
   - Verify token hasn't expired
   - Check token format (Bearer prefix)

3. **Network Errors**
   - Verify backend is running
   - Check API_BASE_URL is correct
   - Check firewall/network settings

4. **Timeout Errors**
   - Increase NEXT_PUBLIC_API_TIMEOUT value
   - Check backend response time
   - Optimize backend queries

## Next Steps

1. Integrate actual authentication with your Go backend
2. Replace mock login with real API calls
3. Implement token refresh logic
4. Add request/response caching
5. Implement retry logic for failed requests
6. Add request cancellation for component unmount
7. Set up API mocking for tests
8. Add request queuing for offline support
9. Implement WebSocket connections if needed
10. Add API monitoring and analytics

