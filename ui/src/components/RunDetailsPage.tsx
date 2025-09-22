import { useState, useEffect } from 'react';
import { useParams, Link as RouterLink } from 'react-router-dom';
import { Typography, CircularProgress, Paper, Box, Breadcrumbs, Link, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';

// Type definitions (can be moved to a shared types file)
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
interface TaskLog {
  id: number;
  task_name: string;
  timestamp: string;
  message: string;
}
interface RunDetails {
  run: PipelineRun;
  logs: TaskLog[];
}

export default function RunDetailsPage() {
  const { id } = useParams<{ id: string }>();
  const [details, setDetails] = useState<RunDetails | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchDetails = async () => {
      try {
        const response = await fetch(`/api/runs/${id}`);
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        setDetails(data);
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

    fetchDetails();
  }, [id]);

  if (loading) {
    return <CircularProgress />;
  }

  if (error) {
    return <Typography color="error">Failed to load run details: {error}</Typography>;
  }

  if (!details) {
    return <Typography>No details found for this run.</Typography>;
  }

  return (
    <Box>
      <Breadcrumbs aria-label="breadcrumb" sx={{ mb: 2 }}>
        <Link component={RouterLink} underline="hover" color="inherit" to="/">
          Runs
        </Link>
        <Typography color="text.primary">Run #{details.run.id}</Typography>
      </Breadcrumbs>

      <Paper sx={{ p: 2, mb: 2 }}>
        <Typography variant="h6" gutterBottom>Run Details</Typography>
        <Typography><b>Group:</b> {details.run.group_name}</Typography>
        <Typography><b>Status:</b> {details.run.status}</Typography>
        <Typography><b>Started:</b> {new Date(details.run.start_time).toLocaleString()}</Typography>
      </Paper>

      <Typography variant="h6" gutterBottom>Logs</Typography>
      <TableContainer component={Paper}>
        <Table sx={{ minWidth: 650 }} size="small">
          <TableHead>
            <TableRow>
              <TableCell>Timestamp</TableCell>
              <TableCell>Task</TableCell>
              <TableCell>Message</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {details.logs.map((log) => (
              <TableRow key={log.id}>
                <TableCell sx={{ whiteSpace: 'nowrap' }}>{new Date(log.timestamp).toLocaleTimeString()}</TableCell>
                <TableCell>{log.task_name}</TableCell>
                <TableCell sx={{ fontFamily: 'monospace', whiteSpace: 'pre-wrap' }}>{log.message}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}
