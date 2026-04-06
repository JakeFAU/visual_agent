// ExecutionEvent is the frontend-normalized shape used by the log and response
// panels, independent of the exact ADK event payload casing.
export interface ExecutionEvent {
  type: string;
  content: any;
  author?: string;
}

export interface UsageSnapshot {
  promptTokens: number;
  responseTokens: number;
  totalTokens: number;
  thoughtsTokens: number;
}

// The backend currently forwards raw ADK events, which may use either Go-style
// exported field names or camelCase names depending on where the payload was
// produced. These helpers normalize both shapes.
const getStateDelta = (content: any): Record<string, unknown> | null => {
  return content?.Actions?.StateDelta ?? content?.actions?.stateDelta ?? null;
};

const getParts = (content: any): any[] => {
  const messageContent = content?.Content ?? content?.content;
  const parts = messageContent?.Parts ?? messageContent?.parts;

  return Array.isArray(parts) ? parts : [];
};

export const getUsageSnapshot = (content: any): UsageSnapshot | null => {
  const usage = content?.UsageMetadata ?? content?.usageMetadata;
  if (!usage) {
    return null;
  }

  return {
    promptTokens: Number(usage.PromptTokenCount ?? usage.promptTokenCount ?? 0),
    responseTokens: Number(usage.CandidatesTokenCount ?? usage.candidatesTokenCount ?? usage.ResponseTokenCount ?? usage.responseTokenCount ?? 0),
    totalTokens: Number(usage.TotalTokenCount ?? usage.totalTokenCount ?? 0),
    thoughtsTokens: Number(usage.ThoughtsTokenCount ?? usage.thoughtsTokenCount ?? 0),
  };
};

const stringifyValue = (value: unknown): string | null => {
  if (typeof value === 'string') {
    return value;
  }

  if (value == null) {
    return null;
  }

  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
};

const getTextFromParts = (parts: any[]): string | null => {
  const text = parts
    .map((part) => {
      if (part?.Thought || part?.thought) {
        return '';
      }

      return part?.Text ?? part?.text ?? '';
    })
    .join('');

  return text.trim() ? text : null;
};

// extractDisplayContent prefers explicit output state keys first, then falls
// back to message parts, then finally stringifies the raw payload for logging.
export const extractDisplayContent = (event: ExecutionEvent, outputKeys: string[] = []): string | null => {
  const stateDelta = getStateDelta(event.content);
  const prioritizedKeys = [...outputKeys, 'message', 'result'];

  for (const key of prioritizedKeys) {
    const value = stateDelta?.[key];
    const text = stringifyValue(value);

    if (text && text.trim()) {
      return text;
    }
  }

  const parts = getParts(event.content);
  const partText = getTextFromParts(parts);
  if (partText) {
    return partText;
  }

  if (stateDelta && Object.keys(stateDelta).length > 0) {
    return stringifyValue(stateDelta);
  }

	return stringifyValue(event.content);
};

// extractFinalResponse is stricter than extractDisplayContent: it only returns
// content appropriate for the "final answer" panel.
export const extractFinalResponse = (event: ExecutionEvent, outputKeys: string[] = []): string | null => {
  const stateDelta = getStateDelta(event.content);
  const prioritizedKeys = [...outputKeys, 'message', 'result'];

  for (const key of prioritizedKeys) {
    const value = stateDelta?.[key];
    const text = stringifyValue(value);

    if (text && text.trim()) {
      return text;
    }
  }

	return getTextFromParts(getParts(event.content));
};

// isFinalAgentResponse filters out partial ADK events so the UI only snapshots
// complete model responses in the final-response panel.
export const isFinalAgentResponse = (event: ExecutionEvent): boolean => {
  if (event.type !== 'agent_event') {
    return false;
  }

  return (event.content?.Partial ?? event.content?.partial) !== true;
};
