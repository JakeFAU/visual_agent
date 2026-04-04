export const API_BASE = 'http://localhost:8080/api';

export const saveGraph = async (graph: any) => {
  const resp = await fetch(`${API_BASE}/graphs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(graph),
  });
  return resp.json();
};

export const loadGraphs = async () => {
  const resp = await fetch(`${API_BASE}/graphs`);
  return resp.json();
};

export const executeGraph = async (graph: any, input: string, onEvent: (ev: { type: string, content: any, author?: string }) => void) => {
  const response = await fetch(`${API_BASE}/execute`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ graph, input }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to start execution');
  }

  const reader = response.body?.getReader();
  if (!reader) return;

  const decoder = new TextDecoder();
  let buffer = '';

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    
    buffer += decoder.decode(value, { stream: true });
    
    // SSE messages are separated by double newlines
    const parts = buffer.split('\n\n');
    
    // The last part might be incomplete, keep it in the buffer
    buffer = parts.pop() || '';

    for (const part of parts) {
        const lines = part.split('\n');
        for (const line of lines) {
            if (line.startsWith('data: ')) {
                try {
                    const jsonStr = line.substring(6).trim();
                    if (jsonStr) {
                        const data = JSON.parse(jsonStr);
                        onEvent(data);
                    }
                } catch (e) {
                    console.warn("Failed to parse SSE JSON:", line, e);
                }
            }
        }
    }
  }
};
