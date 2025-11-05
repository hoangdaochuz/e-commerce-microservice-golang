# Healthcare Portal - Customer App

A modern healthcare portal built with Next.js 16, TypeScript, and Tailwind CSS. This application allows patients to manage their health records, appointments, and prescriptions.

## Features

✨ **Patient Dashboard**
- Home page with personalized greeting and patient avatar
- Feature cards for Medical Records, Appointments, and Prescriptions
- Beautiful gradient design with dark mode support

👤 **Patient Profile**
- Comprehensive profile management
- Editable personal information and address
- Medical history, allergies, and medications tracking
- Avatar display and upload capability

🔐 **Authentication**
- Secure login and registration
- JWT token management
- Session persistence with localStorage
- Protected routes

🌐 **API Integration**
- Axios-based HTTP client with interceptors
- Environment-based configuration
- Automatic error handling
- Request/response logging for development

## Tech Stack

- **Framework:** Next.js 16 (React 19)
- **Language:** TypeScript 5
- **Styling:** Tailwind CSS 4
- **HTTP Client:** Axios
- **State Management:** React Context API
- **Package Manager:** Bun (or npm/yarn/pnpm)

## Getting Started

### Prerequisites

- Node.js 18+ or Bun 1.0+
- Go backend API running (default: http://localhost:8080)

### Installation

1. **Clone the repository**

```bash
cd frontend/customer_app
```

2. **Install dependencies**

```bash
bun install
# or
npm install
```

3. **Set up environment variables**

Copy the example environment file and configure it:

```bash
cp .env.example .env.local
```

Edit `.env.local` with your configuration:

```env
# API Configuration
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_API_TIMEOUT=30000

# Authentication
NEXT_PUBLIC_AUTH_TOKEN_KEY=auth_token

# Feature Flags
NEXT_PUBLIC_ENABLE_LOGGING=true
```

4. **Run the development server**

```bash
bun dev
# or
npm run dev
```

5. **Open your browser**

Navigate to [http://localhost:3000](http://localhost:3000)

## Project Structure

```
frontend/customer_app/
├── app/                    # Next.js app directory
│   ├── layout.tsx         # Root layout with AuthProvider
│   ├── page.tsx           # Home page
│   ├── profile/           # Profile page
│   └── globals.css        # Global styles
├── components/            # Reusable components
│   └── Header.tsx         # Navigation header
├── contexts/              # React contexts
│   └── AuthContext.tsx    # Authentication context
├── lib/                   # Utilities and configurations
│   └── axios.ts           # Axios configuration
├── services/              # API service layer
│   ├── authService.ts     # Authentication APIs
│   ├── patientService.ts  # Patient profile APIs
│   ├── appointmentService.ts  # Appointment APIs
│   ├── prescriptionService.ts # Prescription APIs
│   └── index.ts           # Service exports
├── types/                 # TypeScript type definitions
│   └── patient.ts         # Patient data model
├── docs/                  # Documentation
│   └── API_SETUP.md       # API setup guide
├── .env.local            # Local environment variables (gitignored)
├── .env.example          # Environment variables template
└── package.json          # Dependencies and scripts
```

## API Services

The application includes four main API service modules:

### 1. **Auth Service** (`services/authService.ts`)
- Login, Register, Logout
- Get current patient
- Password reset functionality
- Token refresh

### 2. **Patient Service** (`services/patientService.ts`)
- Get and update patient profile
- Upload avatar
- Manage medical history, allergies, medications

### 3. **Appointment Service** (`services/appointmentService.ts`)
- Create, update, cancel appointments
- Get appointment history
- Check available time slots

### 4. **Prescription Service** (`services/prescriptionService.ts`)
- View prescriptions
- Request refills
- Download prescription PDFs

For detailed API documentation, see [API_SETUP.md](docs/API_SETUP.md).

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_BASE_URL` | Backend API base URL | `http://localhost:8080/api/v1` |
| `NEXT_PUBLIC_API_TIMEOUT` | API timeout (ms) | `30000` |
| `NEXT_PUBLIC_AUTH_TOKEN_KEY` | LocalStorage key for token | `auth_token` |
| `NEXT_PUBLIC_ENABLE_LOGGING` | Enable API logging | `true` |

## Development

### Available Scripts

```bash
# Start development server
bun dev

# Build for production
bun run build

# Start production server
bun start

# Run linter
bun run lint
```

### Code Style

- TypeScript for type safety
- ESLint for code quality
- Tailwind CSS for styling
- React Server Components where possible
- Client Components for interactivity

## Documentation

- [API Setup Guide](docs/API_SETUP.md) - Comprehensive API integration guide
- [Implementation Details](IMPLEMENTATION.md) - Feature implementation guide

## Mock Data

For development and testing, the app includes mock login functionality:

**Demo Patient:**
- Name: John Doe
- Email: john.doe@example.com
- Phone: (555) 123-4567
- Date of Birth: May 15, 1990

Click "Sign In" on the home page to use the mock login.

## Next Steps

To integrate with your Go backend:

1. Update `.env.local` with your backend URL
2. Replace mock authentication in `app/page.tsx`
3. Implement real API endpoints matching the service interfaces
4. Add proper error handling and validation
5. Implement token refresh logic
6. Add unit and integration tests

## Contributing

1. Create a feature branch
2. Make your changes
3. Test thoroughly
4. Submit a pull request

## Learn More

- [Next.js Documentation](https://nextjs.org/docs)
- [Axios Documentation](https://axios-http.com/docs/intro)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [TypeScript Documentation](https://www.typescriptlang.org/docs)

## License

This project is part of the e-commerce microservice golang system.
