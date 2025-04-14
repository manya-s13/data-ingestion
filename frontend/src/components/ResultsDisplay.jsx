import React from 'react';
import { Card, Alert, Button } from 'react-bootstrap';

function ResultsDisplay({ results }) {
  if (!results) {
    return (
      <Card>
        <Card.Header as="h5">Ingestion Results</Card.Header>
        <Card.Body>
          <Alert variant="info">
            No results available. Please run an ingestion job first.
          </Alert>
        </Card.Body>
      </Card>
    );
  }

  return (
    <Card>
      <Card.Header as="h5">Ingestion Results</Card.Header>
      <Card.Body>
        <Alert variant={results.status === 'success' ? 'success' : 'danger'}>
          <Alert.Heading>
            {results.status === 'success' ? 'Success!' : 'Error'}
          </Alert.Heading>
          <p>{results.message}</p>
          {results.status === 'success' && (
            <p className="mb-0">
              <strong>Records processed:</strong> {results.recordCount}
            </p>
          )}
        </Alert>
        
        {results.details && (
          <div className="mt-3">
            <h6>Additional Details:</h6>
            <pre className="bg-light p-3 rounded">
              {JSON.stringify(results.details, null, 2)}
            </pre>
          </div>
        )}
        
        <div className="d-flex justify-content-end mt-3">
          <Button variant="primary" onClick={() => window.location.reload()}>
            Start New Ingestion
          </Button>
        </div>
      </Card.Body>
    </Card>
  );
}

export default ResultsDisplay;