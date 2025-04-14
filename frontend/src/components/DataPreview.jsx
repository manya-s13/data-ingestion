import React from 'react';
import { Card, Table, Button, Spinner, Alert } from 'react-bootstrap';

function DataPreview({ data, loading, onStartIngestion }) {
  if (loading) {
    return (
      <Card>
        <Card.Header as="h5">Data Preview</Card.Header>
        <Card.Body className="text-center py-5">
          <Spinner animation="border" role="status">
            <span className="visually-hidden">Loading...</span>
          </Spinner>
          <p className="mt-3">Loading preview data...</p>
        </Card.Body>
      </Card>
    );
  }

  if (!data || !data.columns || !data.data || data.data.length === 0) {
    return (
      <Card>
        <Card.Header as="h5">Data Preview</Card.Header>
        <Card.Body>
          <Alert variant="info">
            No preview data available. Please configure your ingestion settings and click "Preview Data".
          </Alert>
        </Card.Body>
      </Card>
    );
  }

  return (
    <Card>
      <Card.Header as="h5">Data Preview (First 100 Records)</Card.Header>
      <Card.Body>
        <div className="table-responsive preview-container mb-3">
          <Table striped bordered hover>
            <thead>
              <tr>
                {data.columns.map((col, idx) => (
                  <th key={idx}>{col}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {data.data.map((row, rowIdx) => (
                <tr key={rowIdx}>
                  {data.columns.map((col, colIdx) => (
                    <td key={colIdx}>{row[col]}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </Table>
        </div>
        
        <div className="d-flex justify-content-between align-items-center">
          <div>
            <small className="text-muted">
              Showing {data.data.length} of {data.totalCount || '?'} records
            </small>
          </div>
          <Button variant="primary" onClick={onStartIngestion}>
            Start Ingestion
          </Button>
        </div>
      </Card.Body>
    </Card>
  );
}

export default DataPreview;
