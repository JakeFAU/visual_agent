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

export const executeGraph = async (graph: any, input: string, onEvent: (ev: any) => void) => {
  const response = await fetch(`${API_BASE}/execute`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ graph, input }),
  });

  const reader = response.body?.getReader();
  if (!reader) return;

  const decoder = new TextDecoder();
  let buffer = '';

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    
    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split('\n');
    
    // Keep the last partial line in the buffer
    buffer = lines.pop() || '';

    for (const line of lines) {
        if (line.trim() === '') continue;
        if (line.startsWith('data: ')) {
            try {
                const jsonStr = line.substring(6).trim();
                if (jsonStr) {
                    const data = JSON.parse(jsonStr);
                    onEvent(data);
                }
            } catch (e) {
                console.warn("Failed to parse SSE line:", line, e);
            }
        }
    }
  }
};
