import { useState, useEffect } from 'react';
import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, Chip } from '@mui/material';
import { Link } from 'react-router-dom';

// A simple type definition for a pipeline run
interface PipelineRun {
  id: number;
  group_name: string;
  status: 'success' | 'failed' | 'running';
  start_time: string;
  end_time: string | null;
}

// TODO: Replace with a real API call
const mockRuns: PipelineRun[] = [
  { id: 1, group_name: 'docker-build-pipeline', status: 'success', start_time: '2025-09-21T21:50:42Z', end_time: '2025-09-21T21:51:05Z' },
  { id: 2, group_name: 'terraform-lifecycle', status: 'failed', start_time: '2025-09-21T22:02:38Z', end_time: '2025-09-21T22:03:00Z' },
  { id: 3, group_name: 'ci-pipeline', status: 'running', start_time: '2025-09-21T22:10:00Z', end_time: null },
];

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
  // const [error, setError] = useState<string | null>(null); // Keep for future API call error handling

  useEffect(() => {
    // In the future, we'll fetch from /api/runs
    // For now, we use mock data.
    setRuns(mockRuns);
  }, []);

  // if (error) {
  //   return <Typography color="error">Failed to load runs: {error}</Typography>;
  // }

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
                {run.end_time ? `${((new Date(run.end_time).getTime() - new Date(run.start_time).getTime()) / 1000).toFixed(2)}s` : '-'}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}
