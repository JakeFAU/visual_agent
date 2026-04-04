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
  console.log("[DEBUG] Starting executeGraph fetch...");
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
  if (!reader) {
    console.error("[DEBUG] No response body reader available");
    return;
  }

  const decoder = new TextDecoder();
  let buffer = '';

  const processLine = (line: string) => {
    console.log("[DEBUG] SSE Raw Line:", line);
    if (line.startsWith('data: ')) {
        try {
            const jsonStr = line.substring(6).trim();
            if (jsonStr) {
                const data = JSON.parse(jsonStr);
                onEvent(data);
            }
        } catch (e) {
            console.warn("[DEBUG] Failed to parse SSE JSON:", line, e);
        }
    }
  };

  while (true) {
    const { value, done } = await reader.read();
    if (done) {
        console.log("[DEBUG] Stream reader done. Remaining buffer:", buffer);
        if (buffer) {
            buffer.split('\n').forEach(processLine);
        }
        break;
    }
    
    const chunk = decoder.decode(value, { stream: true });
    console.log("[DEBUG] Received chunk length:", chunk.length);
    buffer += chunk;
    
    const lines = buffer.split('\n');
    // Keep the last partial line
    buffer = lines.pop() || '';

    lines.forEach(processLine);
  }
};
