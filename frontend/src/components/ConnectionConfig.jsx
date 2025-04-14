import React, { useState, useEffect } from 'react';
import { Card, Form, Button, Alert, Tabs, Tab } from 'react-bootstrap';
import api from '../services/api';

function ConnectionConfig({ config, onUpdate, token }) {
  const [currentConfig, setCurrentConfig] = useState(config);
  const [testStatus, setTestStatus] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('clickhouse');

  const handleInputChange = (e) => {
    const { id, value } = e.target;
    setCurrentConfig({
      ...currentConfig,
      [id]: value
    });
  };

  const testConnection = async () => {
    setIsLoading(true);
    setTestStatus(null);
    
    try {
      // Call API to test connection
      await api.testConnection(currentConfig, token);
      setTestStatus({ success: true, message: 'Connection successful!' });
    } catch (error) {
      setTestStatus({ success: false, message: `Connection failed: ${error.message}` });
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    onUpdate(currentConfig);
  };

  return (
    <Card>
      <Card.Header as="h5">Connection Configuration</Card.Header>
      <Card.Body>
        <Tabs 
          activeKey={activeTab}
          onSelect={(k) => setActiveTab(k)}
          className="mb-3"
        >
          <Tab eventKey="clickhouse" title="ClickHouse Connection">
            <Form onSubmit={handleSubmit}>
              <Form.Group className="mb-3">
                <Form.Label>Host</Form.Label>
                <Form.Control 
                  type="text" 
                  id="host" 
                  value={currentConfig.host}
                  onChange={handleInputChange}
                  placeholder="localhost" 
                />
              </Form.Group>
              
              <Form.Group className="mb-3">
                <Form.Label>Port</Form.Label>
                <Form.Control 
                  type="text" 
                  id="port" 
                  value={currentConfig.port}
                  onChange={handleInputChange}
                  placeholder="9000" 
                />
              </Form.Group>
              
              <Form.Group className="mb-3">
                <Form.Label>Database</Form.Label>
                <Form.Control 
                  type="text" 
                  id="database" 
                  value={currentConfig.database}
                  onChange={handleInputChange}
                  placeholder="default" 
                />
              </Form.Group>
              
              <Form.Group className="mb-3">
                <Form.Label>Username</Form.Label>
                <Form.Control 
                  type="text" 
                  id="username" 
                  value={currentConfig.username}
                  onChange={handleInputChange}
                  placeholder="default" 
                />
              </Form.Group>
              
              <Form.Group className="mb-3">
                <Form.Label>Password</Form.Label>
                <Form.Control 
                  type="password" 
                  id="password" 
                  value={currentConfig.password}
                  onChange={handleInputChange}
                  placeholder="Password" 
                />
              </Form.Group>

              <Form.Group className="mb-3">
                <Form.Label>JWT Token</Form.Label>
                <Form.Control 
                  type="password" 
                  id="jwtToken" 
                  value={currentConfig.jwtToken}
                  onChange={handleInputChange}
                  placeholder="JWT Token for authentication" 
                />
                <Form.Text className="text-muted">
                  JWT token for ClickHouse authentication
                </Form.Text>
              </Form.Group>
              
              {testStatus && (
                <Alert variant={testStatus.success ? 'success' : 'danger'}>
                  {testStatus.message}
                </Alert>
              )}
              
              <div className="d-flex gap-2">
                <Button 
                  variant="secondary" 
                  onClick={testConnection} 
                  disabled={isLoading}
                >
                  {isLoading ? 'Testing...' : 'Test Connection'}
                </Button>
                <Button 
                  variant="primary" 
                  type="submit"
                >
                  Save & Continue
                </Button>
              </div>
            </Form>
          </Tab>
        </Tabs>
      </Card.Body>
    </Card>
  );
}

export default ConnectionConfig;
