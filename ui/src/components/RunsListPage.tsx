import { useState, useEffect } from 'react';
import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, Chip, Typography, CircularProgress } from '@mui/material';
import { Link } from 'react-router-dom';

// A simple type definition for a pipeline run
interface PipelineRun {
  id: number;
  group_name: string;
  status: 'success' | 'failed' | 'running';
  start_time: string;
  end_time: {
    Time: string;
    Valid: boolean;
  } | null;
}

const StatusChip = ({ status }: { status: PipelineRun['status'] }) => {
  let color: 'success' | 'error' | 'info' = 'info';
  if (status === 'success') {
    color = 'success';
  } else if (status === 'failed') {
    color = 'error';
  }
  return <Chip label={status} color={color} size="small" />;
};


export default function RunsListPage() {
  const [runs, setRuns] = useState<PipelineRun[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchRuns = async () => {
      try {
        const response = await fetch('/api/runs');
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        setRuns(data || []);
      } catch (e) {
        if (e instanceof Error) {
          setError(e.message);
        } else {
          setError('An unknown error occurred');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchRuns();
  }, []);

  if (loading) {
    return <CircularProgress />;
  }

  if (error) {
    return <Typography color="error">Failed to load runs: {error}</Typography>;
  }

  return (
    <TableContainer component={Paper}>
      <Table sx={{ minWidth: 650 }} aria-label="pipeline runs table">
        <TableHead>
          <TableRow>
            <TableCell>ID</TableCell>
            <TableCell>Group Name</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>Start Time</TableCell>
            <TableCell>Duration</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {runs.map((run) => (
            <TableRow
              key={run.id}
              sx={{ '&:last-child td, &:last-child th': { border: 0 }, 'textDecoration': 'none' }}
              hover
              component={Link} to={`/runs/${run.id}`}
            >
              <TableCell component="th" scope="row">
                {run.id}
              </TableCell>
              <TableCell>{run.group_name}</TableCell>
              <TableCell><StatusChip status={run.status} /></TableCell>
              <TableCell>{new Date(run.start_time).toLocaleString()}</TableCell>
              <TableCell>
                {run.end_time && run.end_time.Valid ? `${((new Date(run.end_time.Time).getTime() - new Date(run.start_time).getTime()) / 1000).toFixed(2)}s` : '-'}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}
