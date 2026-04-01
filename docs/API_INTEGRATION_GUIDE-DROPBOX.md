# Biatech Dropbox API Integration Guide

## Overview
This microservice provides a REST API interface for Dropbox file operations. The service acts as a proxy between your frontend application and Dropbox, handling file uploads, downloads, listing, and search operations.

**Base URL**: `https://cloudcalls.easipath.com/backend-biatechdropbox/api`

## Authentication
This service uses internal Dropbox OAuth tokens and doesn't require client-side authentication. The service automatically refreshes access tokens using stored refresh tokens.

## Path Scoping
All file paths are automatically scoped to `/applications/biatech/` directory in Dropbox. When you provide a path parameter, it will be prefixed with the app scope:

- User path: `""` → Dropbox path: `/applications/biatech`
- User path: `documents` → Dropbox path: `/applications/biatech/documents`  
- User path: `/reports/monthly` → Dropbox path: `/applications/biatech/reports/monthly`

This ensures the service only operates within the app's designated folder structure.

## API Endpoints

### 1. File Upload
Upload files to Dropbox through the service.

**Endpoint**: `POST /upload`

**Parameters**:
- `path` (query parameter, optional): Destination path in Dropbox. If not provided, defaults to `/uploaded-file.tmp`

**Request**:
- Content-Type: `multipart/form-data`
- Form field: `file` (the file to upload)

**Example Request**:
```bash
curl -X POST \
  "https://cloudcalls.easipath.com/backend-biatechdropbox/api/upload?path=documents/report.pdf" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@/local/path/to/report.pdf"
```

*Note: The file will be uploaded to `/applications/biatech/documents/report.pdf` in Dropbox.*

**JavaScript Example**:
```javascript
const uploadFile = async (file, destinationPath) => {
  const formData = new FormData();
  formData.append('file', file);
  
  const response = await fetch(`https://cloudcalls.easipath.com/backend-biatechdropbox/api/upload?path=${encodeURIComponent(destinationPath)}`, {
    method: 'POST',
    body: formData
  });
  
  if (response.ok) {
    const result = await response.text();
    console.log('Upload successful:', result);
  } else {
    throw new Error('Upload failed');
  }
};
```

**Response**:
- **Success (200)**: Plain text message indicating successful upload
- **Error (400/500)**: Error message describing the failure

### 2. List Files
Retrieve a list of files and folders from a Dropbox directory.

**Endpoint**: `GET /list`

**Parameters**:
- `path` (query parameter, optional): Folder path to list. Defaults to root folder if not provided

**Example Request**:
```bash
curl "https://cloudcalls.easipath.com/backend-biatechdropbox/api/list?path=documents"
```

*Note: This will list contents of `/applications/biatech/documents/` in Dropbox.*

**JavaScript Example**:
```javascript
const listFiles = async (folderPath = '') => {
  const response = await fetch(`https://cloudcalls.easipath.com/backend-biatechdropbox/api/list?path=${encodeURIComponent(folderPath)}`);
  
  if (response.ok) {
    const files = await response.json();
    return files;
  } else {
    throw new Error('Failed to list files');
  }
};
```

**Response**:
- **Success (200)**: JSON object with Dropbox API response containing file/folder metadata
- **Error (500)**: Error message

**Example Response Structure**:
```json
{
  "entries": [
    {
      ".tag": "file",
      "name": "document.pdf",
      "path_lower": "/documents/document.pdf",
      "path_display": "/documents/document.pdf",
      "id": "id:abc123",
      "client_modified": "2024-03-22T10:00:00Z",
      "server_modified": "2024-03-22T10:00:00Z",
      "rev": "rev123",
      "size": 1024
    },
    {
      ".tag": "folder",
      "name": "subfolder",
      "path_lower": "/documents/subfolder",
      "path_display": "/documents/subfolder",
      "id": "id:def456"
    }
  ],
  "cursor": "cursor_value",
  "has_more": false
}
```

### 3. Stream File by Revision
Download/stream a file from Dropbox using its revision ID.

**Endpoint**: `GET /stream/{rev}`

**Parameters**:
- `rev` (path parameter, required): The revision ID of the file to download

**Example Request**:
```bash
curl "https://cloudcalls.easipath.com/backend-biatechdropbox/api/stream/rev123abc" \
  -o downloaded_file.pdf
```

**JavaScript Example**:
```javascript
const downloadFile = async (revisionId, filename) => {
  const response = await fetch(`https://cloudcalls.easipath.com/backend-biatechdropbox/api/stream/${revisionId}`);
  
  if (response.ok) {
    const blob = await response.blob();
    
    // Create download link
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    window.URL.revokeObjectURL(url);
  } else {
    throw new Error('Failed to download file');
  }
};
```

**Response**:
- **Success (200)**: File content stream with appropriate content-type header
- **Error (400/404/500)**: Error message

### 4. Search Files
Search for files and folders in Dropbox.

**Endpoint**: `GET /search`

**Parameters**:
- `query` (query parameter, required): Search query string
- `path` (query parameter, optional): Folder path to search within. Defaults to root if not provided

**Example Request**:
```bash
curl "https://cloudcalls.easipath.com/backend-biatechdropbox/api/search?query=report&path=documents"
```

*Note: This will search within `/applications/biatech/documents/` in Dropbox.*

**JavaScript Example**:
```javascript
const searchFiles = async (searchQuery, searchPath = '') => {
  const params = new URLSearchParams({
    query: searchQuery,
    ...(searchPath && { path: searchPath })
  });
  
  const response = await fetch(`https://cloudcalls.easipath.com/backend-biatechdropbox/api/search?${params}`);
  
  if (response.ok) {
    const results = await response.json();
    return results;
  } else {
    throw new Error('Search failed');
  }
};
```

**Response**:
- **Success (200)**: JSON object with Dropbox search results
- **Error (400/500)**: Error message

**Example Response Structure**:
```json
{
  "matches": [
    {
      "metadata": {
        ".tag": "metadata",
        "metadata": {
          ".tag": "file",
          "name": "quarterly_report.pdf",
          "path_lower": "/documents/quarterly_report.pdf",
          "path_display": "/documents/quarterly_report.pdf",
          "id": "id:xyz789",
          "rev": "rev456"
        }
      }
    }
  ],
  "more": false,
  "start": 0
}
```

## Frontend Integration Examples

### React Hook for File Operations
```javascript
import { useState } from 'react';

const useDropboxAPI = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  const baseURL = 'https://cloudcalls.easipath.com/backend-biatechdropbox/api';
  
  const uploadFile = async (file, path) => {
    setLoading(true);
    setError(null);
    try {
      const formData = new FormData();
      formData.append('file', file);
      
      const response = await fetch(`${baseURL}/upload?path=${encodeURIComponent(path)}`, {
        method: 'POST',
        body: formData
      });
      
      if (!response.ok) throw new Error('Upload failed');
      return await response.text();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };
  
  const listFiles = async (path = '') => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(`${baseURL}/list?path=${encodeURIComponent(path)}`);
      if (!response.ok) throw new Error('Failed to list files');
      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };
  
  const downloadFile = async (revisionId, filename) => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(`${baseURL}/stream/${revisionId}`);
      if (!response.ok) throw new Error('Download failed');
      
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };
  
  const searchFiles = async (query, path = '') => {
    setLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams({ query, ...(path && { path }) });
      const response = await fetch(`${baseURL}/search?${params}`);
      if (!response.ok) throw new Error('Search failed');
      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };
  
  return {
    uploadFile,
    listFiles,
    downloadFile,
    searchFiles,
    loading,
    error
  };
};

export default useDropboxAPI;
```

### Vue.js Composable
```javascript
import { ref } from 'vue';

export function useDropboxAPI() {
  const loading = ref(false);
  const error = ref(null);
  
  const baseURL = 'https://cloudcalls.easipath.com/backend-biatechdropbox/api';
  
  const uploadFile = async (file, path) => {
    loading.value = true;
    error.value = null;
    try {
      const formData = new FormData();
      formData.append('file', file);
      
      const response = await fetch(`${baseURL}/upload?path=${encodeURIComponent(path)}`, {
        method: 'POST',
        body: formData
      });
      
      if (!response.ok) throw new Error('Upload failed');
      return await response.text();
    } catch (err) {
      error.value = err.message;
      throw err;
    } finally {
      loading.value = false;
    }
  };
  
  // ... other methods similar to React hook
  
  return {
    uploadFile,
    listFiles,
    downloadFile,
    searchFiles,
    loading,
    error
  };
}
```

## Error Handling
All endpoints return appropriate HTTP status codes:
- **200**: Success
- **400**: Bad request (missing parameters, invalid file format, etc.)
- **500**: Internal server error (Dropbox API errors, server issues, etc.)

Error responses typically include descriptive error messages in the response body.

## Rate Limiting
This service doesn't implement additional rate limiting beyond what Dropbox API enforces. Be mindful of Dropbox API rate limits when making frequent requests.

## Security Notes
- The service handles Dropbox authentication internally
- No client credentials are exposed
- File paths should be properly encoded when passed as query parameters
- Consider implementing client-side file validation before upload

## Testing Results & Known Issues

### Current Status
The service has been thoroughly tested both locally and in production deployment.

**Local Test Results** (Working):
- ✅ `GET /list` - **SUCCESS** - Returns proper directory structure
- ✅ `GET /search` - **SUCCESS** - Returns search results  
- ✅ `POST /upload` - **SUCCESS** - Files uploaded successfully
- ✅ `GET /stream/{rev}` - **SUCCESS** - File download works perfectly

**Live Deployment Status**:
- Deployment configured through docker-compose.yml with .env credentials
- Service code fully functional and ready for production use

### OAuth Scope Issues
The current Dropbox app configuration is missing required API scopes:

**Required Scopes**:
- `files.metadata.read` - For listing and searching files
- `files.content.read` - For downloading/streaming files  
- `files.content.write` - For uploading files

### Error Response Formats

**Missing Scope Error**:
```json
{
  "error": {
    ".tag": "missing_scope",
    "required_scope": "files.metadata.read"
  },
  "error_summary": "missing_scope/"
}
```

**Expired Token Error**:
```json
{
  "error": {
    ".tag": "expired_access_token"
  },
  "error_summary": "expired_access_token/"
}
```

### Troubleshooting

#### For Missing Scope Errors:
1. **Update Dropbox App Permissions**: Go to Dropbox App Console and enable required scopes:
   - `files.metadata.read`
   - `files.content.read` 
   - `files.content.write`
2. **Regenerate Tokens**: After updating scopes, regenerate the refresh token
3. **Update Environment Variables**: Set the new refresh token in your environment
4. **Restart Service**: Restart to pick up new configuration

#### For Expired Token Errors:
1. **Check Token Refresh**: The service now automatically refreshes tokens on each request
2. **Verify Environment Variables**: Ensure `APP_KEY`, `APP_SECRET`, and `REFRESH_TOKEN` are properly set
3. **Check Logs**: Monitor service logs for token refresh failures

### Testing
The service includes e2e tests demonstrating file upload functionality. Test files are available in the `/e2e_test` directory.

Example test URL from codebase:
```
https://cloudcalls.easipath.com/backend-biatechdropbox/api/upload?path=/companies/pmis/docs/payslip/C100006/example.pdf
```

### Manual Testing Commands
Once the token issue is resolved, you can test endpoints with:

```bash
# Test file listing
curl "https://cloudcalls.easipath.com/backend-biatechdropbox/api/list"

# Test file search
curl "https://cloudcalls.easipath.com/backend-biatechdropbox/api/search?query=test"

# Test file upload
curl -X POST "https://cloudcalls.easipath.com/backend-biatechdropbox/api/upload?path=/test_file.txt" \
  -F "file=@/path/to/your/file.txt"
```