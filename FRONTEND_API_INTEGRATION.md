# Frontend API Integration Guide

This document describes the API integration features added to the frontend to work with the tile template backend service.

## ✨ New Features Added

### 🔌 **API Service Layer** (`src/services/`)

#### 1. **API Client** (`src/services/api.ts`)
- Full-featured HTTP client for backend communication
- Automatic error handling and type conversion
- Support for all backend endpoints:
  - `POST /templates` - Create templates
  - `GET /templates` - List templates with search/pagination
  - `GET /templates/{id}` - Get specific template
  - `POST /templates/validate` - Validate templates
  - `GET /health` - Health check

#### 2. **Template Converter** (`src/services/templateConverter.ts`)
- Bi-directional conversion between frontend and backend formats
- Template name validation
- Utility functions for formatting and display

### 💾 **Save/Load System**

#### **SaveLoadPanel Component** (`src/components/new/SaveLoadPanel.tsx`)
- **Save Tab**: Save current template to backend with custom name
- **Load Tab**: Browse, search, and load saved templates
- **Search**: Filter templates by name with live search
- **Error Handling**: User-friendly error messages and retry options

#### **Enhanced Store** (`src/store/newTemplateStore.ts`)
- New API state management (`apiState`)
- Async actions for save/load operations
- Backend validation integration
- Loading states and error handling

### 🎛️ **Updated User Interface**

#### **Enhanced ToolBar**
- 💾 **Save/Load Button**: Opens save/load panel
- 🔍 **Validate Button**: Server-side validation with strict rules
- 📤 **Export JSON**: Enhanced with validation checks
- ⚡ **Loading Indicators**: Visual feedback for API operations

#### **Status Display**
- Real-time API operation status
- Last saved template information
- Error notifications with actionable feedback
- Backend connection status

## 🚀 Quick Setup

### 1. Environment Configuration

Create `.env` file from template:
```bash
cp .env.example .env
```

Update API URL if needed:
```env
VITE_API_BASE_URL=http://localhost:8090/api/v1
```

### 2. Start Backend Service

Ensure the backend is running:
```bash
cd tile-backend
go run cmd/server/main.go
# Server starts on port 8090
```

### 3. Start Frontend

```bash
npm install
npm run dev
# Frontend starts on port 5173
```

## 📖 Usage Guide

### **Saving Templates**

1. Design your template using the layer editors
2. Click **💾 Save/Load** button in toolbar
3. Switch to **Save** tab if not already selected
4. Enter a descriptive template name
5. Click **Save Template**
6. Success confirmation will appear

### **Loading Templates**

1. Click **💾 Save/Load** button in toolbar
2. Switch to **Load** tab
3. Browse available templates or use search
4. Click **Load** button on desired template
5. Template will replace current editor content

### **Backend Validation**

1. Click **🔍 Validate** button in toolbar
2. Template is sent to backend for strict validation
3. Results appear in error summary panel
4. Red borders highlight invalid cells

### **Template Management**

- **Search**: Use name filter to find specific templates
- **Pagination**: Browse large template collections
- **Template Info**: View size, creation date, version
- **Quick Actions**: Load templates with single click

## 🔧 Technical Details

### **API Integration Architecture**

```
Frontend (React + Zustand)
    ↓
API Service Layer (TypeScript)
    ↓
HTTP Client (fetch)
    ↓
Backend REST API (Go + PostgreSQL)
```

### **Data Flow**

1. **Save Operation**:
   ```
   Frontend Template → Type Converter → API Request → Backend → Database
   ```

2. **Load Operation**:
   ```
   Database → Backend → API Response → Type Converter → Frontend Template
   ```

### **Error Handling**

- **Network Errors**: Automatic retry suggestions
- **Validation Errors**: Detailed field-level feedback
- **Server Errors**: User-friendly error messages
- **Connection Issues**: Clear connection status indicators

### **State Management**

```typescript
interface ApiState {
  isLoading: boolean;        // Current operation status
  error: string | null;      // Last error message
  lastSaved?: {              // Last successful save
    id: string;
    name: string;
    savedAt: string;
  };
}
```

## 🎨 UI/UX Enhancements

### **Visual Feedback**
- ⏳ Loading spinners during API calls
- ✅ Success notifications for completed operations
- ❌ Error messages with clear explanations
- 🔄 Real-time status updates

### **User Experience**
- **Auto-generated Names**: Default names with timestamps
- **Search as You Type**: Instant template filtering
- **Visual Template Cards**: Easy browsing with metadata
- **Keyboard Navigation**: Tab support for accessibility

### **Error Recovery**
- **Clear Error States**: Easy error dismissal
- **Retry Mechanisms**: One-click error recovery
- **Validation Feedback**: Precise error locations
- **Connection Monitoring**: Backend health indicators

## 🔍 API Endpoints Used

| Method | Endpoint | Purpose | Frontend Usage |
|--------|----------|---------|----------------|
| `POST` | `/templates` | Create template | Save operation |
| `GET` | `/templates` | List templates | Load panel template list |
| `GET` | `/templates/{id}` | Get template | Load specific template |
| `POST` | `/templates/validate` | Validate template | Backend validation |
| `GET` | `/health` | Health check | Connection status |

## 📋 Configuration Options

### **Environment Variables**

```env
# Required: Backend API base URL
VITE_API_BASE_URL=http://localhost:8090/api/v1

# Optional: Development settings
VITE_NODE_ENV=development
```

### **API Client Configuration**

```typescript
// Custom API instance
const customApi = new TemplateApiService('http://custom-backend:8080/api/v1');

// With custom headers
const authenticatedApi = new TemplateApiService(baseUrl, {
  headers: { 'Authorization': 'Bearer token' }
});
```

## 🚨 Error Handling Examples

### **Network Issues**
```
❌ API Error: Network error: Failed to fetch
→ Check backend connection and try again
```

### **Validation Errors**
```
❌ Validation Failed: Template validation failed
→ Fix highlighted errors and validate again
```

### **Server Issues**
```
❌ API Error: Internal server error
→ Backend service may be down, try again later
```

## 🎯 Best Practices

### **For Users**
1. **Save Early, Save Often**: Regular saves prevent data loss
2. **Use Descriptive Names**: Easy template identification
3. **Validate Before Save**: Catch errors early
4. **Check Connection**: Monitor backend status

### **For Developers**
1. **Error Boundaries**: Wrap API calls in try-catch
2. **Loading States**: Always show operation progress
3. **Type Safety**: Use TypeScript interfaces consistently
4. **User Feedback**: Clear success/error messaging

## 🔮 Future Enhancements

### **Planned Features**
- 🔄 **Auto-save**: Periodic automatic saves
- 📂 **Template Folders**: Organization and categorization  
- 👥 **Collaboration**: Multi-user template sharing
- 📊 **Analytics**: Usage statistics and insights
- 🔒 **Authentication**: User accounts and permissions

### **Technical Improvements**
- **Offline Support**: Local storage fallback
- **Caching**: Client-side template caching
- **Real-time Sync**: WebSocket-based live updates
- **Batch Operations**: Multiple template management

This integration provides a solid foundation for professional template management with excellent user experience and robust error handling! 🚀