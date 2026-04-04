export interface ExecutionEvent {
  type: string;
  content: any;
  author?: string;
}

const getStateDelta = (content: any): Record<string, unknown> | null => {
  return content?.Actions?.StateDelta ?? content?.actions?.stateDelta ?? null;
};

const getParts = (content: any): any[] => {
  const messageContent = content?.Content ?? content?.content;
  const parts = messageContent?.Parts ?? messageContent?.parts;

  return Array.isArray(parts) ? parts : [];
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

export const isFinalAgentResponse = (event: ExecutionEvent): boolean => {
  if (event.type !== 'agent_event') {
    return false;
  }

  return (event.content?.Partial ?? event.content?.partial) !== true;
};
