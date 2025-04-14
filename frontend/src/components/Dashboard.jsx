import React, { useState } from 'react';
import { Container, Row, Col, Nav } from 'react-bootstrap';
import Header from './Header';
import ConnectionConfig from './ConnectionConfig';
import IngestionForm from './IngestionForm';
import DataPreview from './DataPreview';
import ResultsDisplay from './ResultsDisplay';

function Dashboard({ token, onLogout }) {
  const [activeTab, setActiveTab] = useState('config');
  const [connectionConfig, setConnectionConfig] = useState({
    host: 'localhost',
    port: '9000',
    database: 'default',
    username: 'default',
    password: '',
    jwtToken: ''
  });
  const [flatFileConfig, setFlatFileConfig] = useState({
    fileName: '',
    delimiter: ','
  });
  const [direction, setDirection] = useState('ch_to_flat');
  const [selectedTable, setSelectedTable] = useState('');
  const [targetTable, setTargetTable] = useState('');
  const [availableTables, setAvailableTables] = useState([]);
  const [columns, setColumns] = useState([]);
  const [selectedColumns, setSelectedColumns] = useState([]);
  const [joinTables, setJoinTables] = useState([]);
  const [joinKeys, setJoinKeys] = useState([]);
  const [previewData, setPreviewData] = useState(null);
  const [results, setResults] = useState(null);
  const [status, setStatus] = useState('idle'); // idle, loading, success, error

  const handleConfigUpdate = (config) => {
    setConnectionConfig(config);
    setActiveTab('ingestion');
  };

  const handleFlatFileConfigUpdate = (config) => {
    setFlatFileConfig(config);
  };

  const handleDirectionChange = (newDirection) => {
    setDirection(newDirection);
    setSelectedColumns([]);
    setPreviewData(null);
  };

  const handleTableSelect = (tableName) => {
    setSelectedTable(tableName);
    setSelectedColumns([]);
    setPreviewData(null);
  };

  const handleColumnToggle = (columnName) => {
    setSelectedColumns(prev => 
      prev.includes(columnName)
        ? prev.filter(col => col !== columnName)
        : [...prev, columnName]
    );
  };

  const handleSelectAllColumns = () => {
    if (selectedColumns.length === columns.length) {
      setSelectedColumns([]);
    } else {
      setSelectedColumns(columns.map(col => col.name));
    }
  };

  const handleJoinTablesChange = (tables) => {
    setJoinTables(tables);
  };

  const handleJoinKeysChange = (keys) => {
    setJoinKeys(keys);
  };

  const handlePreview = async () => {
    setStatus('loading');
    try {
      // Logic for preview would go here
      // This would call your API endpoint for preview
      setPreviewData({ /* preview data structure */ });
      setStatus('success');
    } catch (error) {
      setStatus('error');
    }
  };

  const handleIngestion = async () => {
    setStatus('loading');
    try {
      // Logic for ingestion would go here
      // This would call your API endpoint for ingestion
      setResults({ /* results data structure */ });
      setStatus('success');
      setActiveTab('results');
    } catch (error) {
      setStatus('error');
    }
  };

  return (
    <div>
      <Header onLogout={onLogout} />
      <Container fluid className="mt-4">
        <Row>
          <Col md={3}>
            <Nav className="flex-column">
              <Nav.Link 
                className={activeTab === 'config' ? 'active' : ''} 
                onClick={() => setActiveTab('config')}
              >
                Connection Configuration
              </Nav.Link>
              <Nav.Link 
                className={activeTab === 'ingestion' ? 'active' : ''} 
                onClick={() => setActiveTab('ingestion')}
              >
                Data Ingestion
              </Nav.Link>
              <Nav.Link 
                className={activeTab === 'preview' ? 'active' : ''} 
                onClick={() => setActiveTab('preview')}
                disabled={!selectedTable || selectedColumns.length === 0}
              >
                Data Preview
              </Nav.Link>
              <Nav.Link 
                className={activeTab === 'results' ? 'active' : ''} 
                onClick={() => setActiveTab('results')}
                disabled={!results}
              >
                Results
              </Nav.Link>
            </Nav>
          </Col>
          <Col md={9}>
            {activeTab === 'config' && (
              <ConnectionConfig 
                config={connectionConfig}
                onUpdate={handleConfigUpdate}
                token={token}
              />
            )}
            
            {activeTab === 'ingestion' && (
              <IngestionForm 
                direction={direction}
                onDirectionChange={handleDirectionChange}
                connectionConfig={connectionConfig}
                flatFileConfig={flatFileConfig}
                onFlatFileConfigUpdate={handleFlatFileConfigUpdate}
                selectedTable={selectedTable}
                onTableSelect={handleTableSelect}
                targetTable={targetTable}
                onTargetTableChange={setTargetTable}
                availableTables={availableTables}
                columns={columns}
                selectedColumns={selectedColumns}
                onColumnToggle={handleColumnToggle}
                onSelectAllColumns={handleSelectAllColumns}
                joinTables={joinTables}
                onJoinTablesChange={handleJoinTablesChange}
                joinKeys={joinKeys}
                onJoinKeysChange={handleJoinKeysChange}
                onPreview={() => {
                  handlePreview();
                  setActiveTab('preview');
                }}
                onIngestion={handleIngestion}
                status={status}
                token={token}
              />
            )}
            
            {activeTab === 'preview' && (
              <DataPreview 
                data={previewData}
                loading={status === 'loading'}
                onStartIngestion={handleIngestion}
              />
            )}
            
            {activeTab === 'results' && (
              <ResultsDisplay results={results} />
            )}
          </Col>
        </Row>
      </Container>
    </div>
  );
}

export default Dashboard;