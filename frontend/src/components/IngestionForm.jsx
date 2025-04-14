import React, { useState, useEffect } from 'react';
import { Card, Form, Button, ListGroup, Alert, Row, Col, Spinner } from 'react-bootstrap';
import api from '../services/api';

function IngestionForm({ 
  direction, 
  onDirectionChange,
  connectionConfig,
  flatFileConfig,
  onFlatFileConfigUpdate,
  selectedTable,
  onTableSelect,
  targetTable,
  onTargetTableChange,
  availableTables,
  columns,
  selectedColumns,
  onColumnToggle,
  onSelectAllColumns,
  joinTables,
  onJoinTablesChange,
  joinKeys,
  onJoinKeysChange,
  onPreview,
  onIngestion,
  status,
  token
}) {
  const [tableList, setTableList] = useState([]);
  const [columnList, setColumnList] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [fileToUpload, setFileToUpload] = useState(null);

  useEffect(() => {
    loadTables();
  }, [connectionConfig]);

  const loadTables = async () => {
    setIsLoading(true);
    setError('');
    
    try {
      const tables = await api.getTables(connectionConfig, token);
      setTableList(tables);
    } catch (err) {
      setError('Failed to load tables: ' + err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const loadColumns = async (tableName) => {
    if (!tableName) return;
    
    setIsLoading(true);
    setError('');
    
    try {
      const columns = await api.getColumns(tableName, connectionConfig, token);
      setColumnList(columns);
    } catch (err) {
      setError('Failed to load columns: ' + err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleTableChange = (e) => {
    const table = e.target.value;
    onTableSelect(table);
    loadColumns(table);
  };

  const handleTargetTableChange = (e) => {
    onTargetTableChange(e.target.value);
  };

  const handleFileChange = (e) => {
    setFileToUpload(e.target.files[0]);
    
    if (e.target.files[0]) {
      onFlatFileConfigUpdate({
        ...flatFileConfig,
        fileName: e.target.files[0].name
      });
    }
  };

  const handleDelimiterChange = (e) => {
    onFlatFileConfigUpdate({
      ...flatFileConfig,
      delimiter: e.target.value
    });
  };

  const handleDirectionRadioChange = (e) => {
    onDirectionChange(e.target.value);
  };

  return (
    <Card>
      <Card.Header as="h5">Data Ingestion Configuration</Card.Header>
      <Card.Body>
        {error && <Alert variant="danger">{error}</Alert>}
        
        <Form.Group className="mb-4">
          <Form.Label>Data Flow Direction</Form.Label>
          <div>
            <Form.Check
              type="radio"
              label="ClickHouse → Flat File"
              name="direction"
              id="ch-to-flat"
              value="ch_to_flat"
              checked={direction === 'ch_to_flat'}
              onChange={handleDirectionRadioChange}
              className="mb-2"
            />
            <Form.Check
              type="radio"
              label="Flat File → ClickHouse"
              name="direction"
              id="flat-to-ch"
              value="flat_to_ch"
              checked={direction === 'flat_to_ch'}
              onChange={handleDirectionRadioChange}
            />
          </div>
        </Form.Group>

        <Row>
          {/* Source Configuration */}
          <Col md={6}>
            <Card className="mb-3">
              <Card.Header>
                {direction === 'ch_to_flat' ? 'ClickHouse Source' : 'Flat File Source'}
              </Card.Header>
              <Card.Body>
                {direction === 'ch_to_flat' ? (
                  <>
                    <Form.Group className="mb-3">
                      <Form.Label>Source Table</Form.Label>
                      <Form.Select 
                        value={selectedTable} 
                        onChange={handleTableChange}
                      >
                        <option value="">Select a table</option>
                        {tableList.map((table, idx) => (
                          <option key={idx} value={table}>{table}</option>
                        ))}
                      </Form.Select>
                    </Form.Group>

                    {/* Bonus: Join Tables */}
                    <Form.Group className="mb-3">
                      <Form.Label>Join Tables (Optional)</Form.Label>
                      <Form.Control 
                        as="textarea" 
                        placeholder="Enter join tables, one per line" 
                        value={joinTables.join('\n')}
                        onChange={(e) => onJoinTablesChange(e.target.value.split('\n').filter(Boolean))}
                      />
                    </Form.Group>

                    <Form.Group className="mb-3">
                      <Form.Label>Join Keys (Optional)</Form.Label>
                      <Form.Control 
                        as="textarea" 
                        placeholder="Format: table1.col = table2.col" 
                        value={joinKeys.join('\n')}
                        onChange={(e) => onJoinKeysChange(e.target.value.split('\n').filter(Boolean))}
                      />
                    </Form.Group>
                  </>
                ) : (
                  <>
                    <Form.Group className="mb-3">
                      <Form.Label>Upload Flat File</Form.Label>
                      <Form.Control 
                        type="file" 
                        onChange={handleFileChange}
                      />
                    </Form.Group>
                    
                    <Form.Group className="mb-3">
                      <Form.Label>Delimiter</Form.Label>
                      <Form.Select 
                        value={flatFileConfig.delimiter} 
                        onChange={handleDelimiterChange}
                      >
                        <option value=",">Comma (,)</option>
                        <option value=";">Semicolon (;)</option>
                        <option value="\t">Tab</option>
                        <option value="|">Pipe (|)</option>
                      </Form.Select>
                    </Form.Group>
                  </>
                )}
              </Card.Body>
            </Card>
          </Col>

          {/* Target Configuration */}
          <Col md={6}>
            <Card className="mb-3">
              <Card.Header>
                {direction === 'ch_to_flat' ? 'Flat File Target' : 'ClickHouse Target'}
              </Card.Header>
              <Card.Body>
                {direction === 'ch_to_flat' ? (
                  <>
                    <Form.Group className="mb-3">
                      <Form.Label>Output File Name</Form.Label>
                      <Form.Control 
                        type="text" 
                        placeholder="output.csv" 
                        value={flatFileConfig.fileName}
                        onChange={(e) => onFlatFileConfigUpdate({...flatFileConfig, fileName: e.target.value})}
                      />
                    </Form.Group>
                    
                    <Form.Group className="mb-3">
                      <Form.Label>Delimiter</Form.Label>
                      <Form.Select 
                        value={flatFileConfig.delimiter} 
                        onChange={handleDelimiterChange}
                      >
                        <option value=",">Comma (,)</option>
                        <option value=";">Semicolon (;)</option>
                        <option value="\t">Tab</option>
                        <option value="|">Pipe (|)</option>
                      </Form.Select>
                    </Form.Group>
                  </>
                ) : (
                  <>
                    <Form.Group className="mb-3">
                      <Form.Label>Target Table</Form.Label>
                      <Form.Control 
                        type="text" 
                        placeholder="Enter target table name" 
                        value={targetTable}
                        onChange={handleTargetTableChange}
                      />
                      <Form.Text className="text-muted">
                        If the table doesn't exist, it will be created
                      </Form.Text>
                    </Form.Group>
                  </>
                )}
              </Card.Body>
            </Card>
          </Col>
        </Row>

        {/* Column Selection */}
        {columnList.length > 0 && (
          <Card className="mb-4">
            <Card.Header>
              <div className="d-flex justify-content-between align-items-center">
                <span>Column Selection</span>
                <Button 
                  variant="outline-primary" 
                  size="sm"
                  onClick={onSelectAllColumns}
                >
                  {selectedColumns.length === columnList.length ? 'Deselect All' : 'Select All'}
                </Button>
              </div>
            </Card.Header>
            <Card.Body>
              <div className="column-selection">
                <ListGroup>
                  {columnList.map((column, idx) => (
                    <ListGroup.Item key={idx} className="d-flex align-items-center">
                      <Form.Check 
                        type="checkbox"
                        id={`column-${idx}`}
                        label={`${column.name} (${column.type})`}
                        checked={selectedColumns.includes(column.name)}
                        onChange={() => onColumnToggle(column.name)}
                      />
                    </ListGroup.Item>
                  ))}
                </ListGroup>
              </div>
            </Card.Body>
          </Card>
        )}

        <div className="d-flex justify-content-end gap-2">
          <Button 
            variant="secondary" 
            onClick={onPreview}
            disabled={
              isLoading || 
              status === 'loading' || 
              !selectedTable || 
              (direction === 'ch_to_flat' && !flatFileConfig.fileName) ||
              (direction === 'flat_to_ch' && !targetTable)
            }
          >
            Preview Data
          </Button>
          <Button 
            variant="primary" 
            onClick={onIngestion}
            disabled={
              isLoading || 
              status === 'loading' || 
              !selectedTable || 
              selectedColumns.length === 0 ||
              (direction === 'ch_to_flat' && !flatFileConfig.fileName) ||
              (direction === 'flat_to_ch' && !targetTable)
            }
          >
            {status === 'loading' ? (
              <>
                <Spinner as="span" animation="border" size="sm" role="status" aria-hidden="true" />
                <span className="ms-2">Processing...</span>
              </>
            ) : 'Start Ingestion'}
          </Button>
        </div>
      </Card.Body>
    </Card>
  );
}

export default IngestionForm;
