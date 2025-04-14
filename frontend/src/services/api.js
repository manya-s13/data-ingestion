// src/services/api.js
import axios from 'axios';

// Base API URL - replace with your actual API endpoint
const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080/api';

// Create axios instance with default config
const axiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  }
});

// Add request interceptor to include auth token
axiosInstance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('authToken');
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// API service object
const api = {
  // Authentication
  login: async (username, password) => {
    try {
      const response = await axiosInstance.post('/auth/login', { username, password });
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // Connection testing
  testConnection: async (connectionConfig, token) => {
    try {
      const response = await axiosInstance.post('/connection/test', connectionConfig, {
        headers: token ? { Authorization: `Bearer ${token}` } : {}
      });
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // Tables operations
  getTables: async (connectionConfig, token) => {
    try {
      const response = await axiosInstance.post('/clickhouse/tables', connectionConfig, {
        headers: token ? { Authorization: `Bearer ${token}` } : {}
      });
      return response.data.tables || [];
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // Columns operations
  getColumns: async (tableName, connectionConfig, token) => {
    try {
      const response = await axiosInstance.post('/clickhouse/columns', {
        tableName,
        ...connectionConfig
      }, {
        headers: token ? { Authorization: `Bearer ${token}` } : {}
      });
      return response.data.columns || [];
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // Data preview
  getDataPreview: async (config, token) => {
    try {
      const response = await axiosInstance.post('/data/preview', config, {
        headers: token ? { Authorization: `Bearer ${token}` } : {}
      });
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // File upload
  uploadFile: async (file, params, token) => {
    try {
      const formData = new FormData();
      formData.append('file', file);
      
      // Add any additional parameters
      Object.keys(params).forEach(key => {
        formData.append(key, params[key]);
      });

      const response = await axiosInstance.post('/upload', formData, {
        headers: {
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
          'Content-Type': 'multipart/form-data'
        }
      });
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // Data ingestion (ClickHouse to Flat File)
  exportToFlatFile: async (config, token) => {
    try {
      const response = await axiosInstance.post('/data/export', config, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        responseType: 'blob' // For file download
      });
      
      // Create file download
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', config.fileName || 'export.csv');
      document.body.appendChild(link);
      link.click();
      link.remove();
      
      return {
        status: 'success',
        message: 'File exported successfully',
        recordCount: response.headers['x-record-count'] || 'Unknown',
        details: {
          fileName: config.fileName || 'export.csv',
          fileSize: (response.data.size / 1024).toFixed(2) + ' KB'
        }
      };
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // Data ingestion (Flat File to ClickHouse)
  importToClickHouse: async (file, config, token) => {
    try {
      const formData = new FormData();
      formData.append('file', file);
      
      // Add configuration parameters
      Object.keys(config).forEach(key => {
        if (typeof config[key] === 'object') {
          formData.append(key, JSON.stringify(config[key]));
        } else {
          formData.append(key, config[key]);
        }
      });

      const response = await axiosInstance.post('/data/import', formData, {
        headers: {
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
          'Content-Type': 'multipart/form-data'
        }
      });
      
      return response.data;
    } catch (error) {
      throw handleApiError(error);
    }
  },

  // Execute data ingestion based on direction
  executeIngestion: async (direction, config, file, token) => {
    if (direction === 'ch_to_flat') {
      return await api.exportToFlatFile(config, token);
    } else {
      return await api.importToClickHouse(file, config, token);
    }
  }
};

// Helper function to handle API errors
function handleApiError(error) {
  let errorMessage = 'An unknown error occurred';
  
  if (error.response) {
    // The server responded with a status code outside of 2xx range
    errorMessage = error.response.data.message || 
                   `Server error: ${error.response.status}`;
  } else if (error.request) {
    // The request was made but no response was received
    errorMessage = 'No response received from server. Please check your network connection.';
  } else {
    // Something happened in setting up the request
    errorMessage = error.message;
  }
  
  const apiError = new Error(errorMessage);
  apiError.originalError = error;
  apiError.status = error.response ? error.response.status : null;
  return apiError;
}

export default api;