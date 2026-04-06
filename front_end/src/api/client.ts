// API_BASE defaults to a same-origin /api prefix so the browser can work both
// behind the Docker nginx reverse proxy and behind the Vite dev proxy.
const configuredAPIBase = import.meta.env.VITE_API_BASE?.trim();

export const API_BASE = (configuredAPIBase && configuredAPIBase.length > 0 ? configuredAPIBase : '/api').replace(/\/$/, '');

// saveGraph persists the currently open graph document to the backend.
export const saveGraph = async (graph: any) => {
  const resp = await fetch(`${API_BASE}/graphs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(graph),
  });

  if (!resp.ok) {
    const error = await resp.json().catch(() => null);
    throw new Error(error?.error || 'Failed to save graph');
  }

	return resp.json();
};

// loadGraphs retrieves the names of saved graphs so the UI can prompt the user.
export const loadGraphs = async () => {
  const resp = await fetch(`${API_BASE}/graphs`);

  if (!resp.ok) {
    const error = await resp.json().catch(() => null);
    throw new Error(error?.error || 'Failed to load graphs');
  }

	return resp.json();
};

// loadGraph fetches a single graph document by its saved name.
export const loadGraph = async (name: string) => {
  const resp = await fetch(`${API_BASE}/graphs/${encodeURIComponent(name)}`);

  if (!resp.ok) {
    const error = await resp.json().catch(() => null);
    throw new Error(error?.error || `Failed to load graph '${name}'`);
  }

	return resp.json();
};

export interface ExecuteBudget {
  max_steps?: number;
  max_duration_ms?: number;
  max_total_tokens?: number;
}

// executeGraph starts workflow execution and forwards the backend's SSE event
// stream to the supplied callback one event at a time.
export const executeGraph = async (
  graph: any,
  input: string,
  budget: ExecuteBudget,
  onEvent: (ev: { type: string, content: any, author?: string }) => void,
) => {
  console.log("[DEBUG] Starting executeGraph fetch...");
  const response = await fetch(`${API_BASE}/execute`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ graph, input, budget }),
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
    const trimmed = line.trim();
    if (!trimmed) return;
    
    console.log("[DEBUG] SSE Raw Line:", trimmed);
    if (trimmed.startsWith('data:')) {
        try {
            // SSE frames arrive as "data: <json>", so only the payload after the
            // first colon should be parsed.
            const jsonStr = trimmed.substring(trimmed.indexOf(':') + 1).trim();
            if (jsonStr) {
                const data = JSON.parse(jsonStr);
                onEvent(data);
            }
        } catch (e) {
            console.warn("[DEBUG] Failed to parse SSE JSON:", trimmed, e);
        }
    }
  };

  while (true) {
    let value: Uint8Array | undefined;
    let done = false;
    try {
      ({ value, done } = await reader.read());
    } catch (e) {
      const message = e instanceof Error ? e.message : String(e);
      throw new Error(`Execution stream closed unexpectedly: ${message}`);
    }
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
    // Keep the last partial line so JSON frames split across chunks can be
    // reassembled on the next read.
    buffer = lines.pop() || '';

    lines.forEach(processLine);
  }
};
