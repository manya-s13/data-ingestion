// src/context/AuthContext.jsx
import React, { createContext, useState, useContext, useEffect } from 'react';
import api from '../services/api';

// Create the authentication context
const AuthContext = createContext(null);

// Custom hook to use the auth context
export const useAuth = () => {
  return useContext(AuthContext);
};

// Provider component
export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [token, setToken] = useState(localStorage.getItem('authToken'));
  const [loading, setLoading] = useState(true);

  // Check if there's a token in local storage on component mount
  useEffect(() => {
    if (token) {
      // Here you could validate the token with your backend
      // For now, we'll just set the current user based on the token existence
      setCurrentUser({ token });
    }
    setLoading(false);
  }, [token]);

  // Login function
  const login = async (username, password) => {
    try {
      const response = await api.login(username, password);
      const { token, user } = response;
      
      // Store token in localStorage
      localStorage.setItem('authToken', token);
      
      // Update state
      setToken(token);
      setCurrentUser(user);
      
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  };

  // Logout function
  const logout = () => {
    // Clear local storage
    localStorage.removeItem('authToken');
    
    // Reset state
    setToken(null);
    setCurrentUser(null);
  };

  // Context value
  const value = {
    currentUser,
    token,
    login,
    logout,
    isAuthenticated: !!currentUser
  };

  return (
    <AuthContext.Provider value={value}>
      {!loading && children}
    </AuthContext.Provider>
  );
};

export default AuthContext;