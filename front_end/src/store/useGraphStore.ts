import { create } from 'zustand';
import {
  Connection,
  Edge,
  EdgeChange,
  Node,
  NodeChange,
  addEdge,
  applyNodeChanges,
  applyEdgeChanges,
  OnNodesChange,
  OnEdgesChange,
  OnConnect,
} from 'reactflow';
import { GraphSchema, Graph, Node as ContractNode, Edge as ContractEdge } from '../schema/graph';

const DEFAULT_NODE_WIDTH = 180;
const DEFAULT_NODE_HEIGHT = 96;
const DEFAULT_WHILE_WIDTH = 460;
const DEFAULT_WHILE_HEIGHT = 300;
const CONTAINER_DROP_PADDING_X = 24;
const CONTAINER_DROP_PADDING_Y = 92;

const legacyPortAliases: Record<string, string> = {
  'message:in': 'message',
  'message:out': 'message',
  'message:enter': 'message',
  in_message: 'message',
  out_message: 'message',
  'toolbox:in': 'toolbox_handle',
  'toolbox:out': 'toolbox_handle',
  in_toolbox: 'toolbox_handle',
  out_true: 'message:true',
  out_false: 'message:false',
  out_loop: 'message:loop',
  out_done: 'message:done',
};

const normalizePort = (port?: string | null) => {
  if (!port) {
    return '';
  }
  return legacyPortAliases[port] ?? port;
};

const toCanvasSourcePort = (nodeType: string | undefined, port?: string | null) => {
  const normalized = normalizePort(port);
  switch (nodeType) {
    case 'input_node':
      return 'message:out';
    case 'llm_node':
      if (normalized === 'toolbox_handle') {
        return 'toolbox:out';
      }
      return 'message:out';
    case 'toolbox':
      return 'toolbox:out';
    case 'if_else_node':
      if (normalized === 'message:true' || normalized === 'message:false') {
        return normalized;
      }
      return normalized || 'message:true';
    case 'while_node':
      if (normalized === 'message:loop' || normalized === 'message:done') {
        return normalized;
      }
      return normalized || 'message:done';
    default:
      return normalized;
  }
};

const toCanvasTargetPort = (nodeType: string | undefined, port?: string | null) => {
  const normalized = normalizePort(port);
  switch (nodeType) {
    case 'llm_node':
      if (normalized === 'toolbox_handle') {
        return 'toolbox:in';
      }
      return 'message:in';
    case 'output_node':
      return 'message:in';
    case 'if_else_node':
      return 'message:in';
    case 'while_node':
      if (normalized === 'message:return') {
        return 'message:return';
      }
      return 'message:enter';
    default:
      return normalized;
  }
};

const preferredMessageInputHandle = (nodeType: string | undefined) => {
  switch (nodeType) {
    case 'llm_node':
      return 'message:in';
    case 'if_else_node':
      return 'message:in';
    case 'output_node':
      return 'message:in';
    case 'while_node':
      return 'message:enter';
    default:
      return 'message:in';
  }
};

const portFamily = (port?: string | null) => {
  const normalized = normalizePort(port);
  if (!normalized) {
    return '';
  }
  if (normalized === 'toolbox_handle') {
    return 'toolbox';
  }
  if (normalized.startsWith('message')) {
    return 'message';
  }
  return normalized.split(':')[0];
};

const handleRole = (nodeType: string | undefined, handle?: string | null): 'source' | 'target' | 'unknown' => {
  switch (nodeType) {
    case 'input_node':
      return handle === 'message:out' ? 'source' : 'unknown';
    case 'llm_node':
      if (handle === 'message:in' || handle === 'toolbox:in') {
        return 'target';
      }
      if (handle === 'message:out') {
        return 'source';
      }
      return 'unknown';
    case 'toolbox':
      return handle === 'toolbox:out' ? 'source' : 'unknown';
    case 'output_node':
      return handle === 'message:in' ? 'target' : 'unknown';
    case 'if_else_node':
      if (handle === 'message:in') {
        return 'target';
      }
      if (handle === 'message:true' || handle === 'message:false') {
        return 'source';
      }
      return 'unknown';
    case 'while_node':
      if (handle === 'message:enter' || handle === 'message:return') {
        return 'target';
      }
      if (handle === 'message:loop' || handle === 'message:done') {
        return 'source';
      }
      return 'unknown';
    default:
      return 'unknown';
  }
};

const normalizeConnectionDirection = (connection: Connection, nodes: Node[]) => {
  const nodeTypes = new Map(nodes.map((node) => [node.id, node.type]));
  const sourceRole = handleRole(nodeTypes.get(connection.source || ''), connection.sourceHandle);
  const targetRole = handleRole(nodeTypes.get(connection.target || ''), connection.targetHandle);

  if (sourceRole === 'target' && targetRole === 'source') {
    return {
      source: connection.target,
      sourceHandle: connection.targetHandle,
      target: connection.source,
      targetHandle: connection.sourceHandle,
    };
  }

  return connection;
};

const isWhileIngressHandle = (handle?: string | null) =>
  handle === 'message:enter' || handle === 'message:return' || handle === 'message:loop';

const normalizeWhileContainerConnection = (connection: Connection, nodes: Node[]) => {
  const nodeIndex = buildNodeIndex(nodes);
  const sourceNode = connection.source ? nodeIndex.get(connection.source) : undefined;
  const targetNode = connection.target ? nodeIndex.get(connection.target) : undefined;

  if (!sourceNode || !targetNode) {
    return connection;
  }

  // If a user wires the visible loop-container ingress points directly to a
  // child node, treat that as "start the body here" and snap the child end to
  // its input handle. This preserves the intended visual flow without
  // rewriting legitimate child-output -> repeat return edges.
  if (
    sourceNode.type === 'while_node' &&
    isWhileIngressHandle(connection.sourceHandle) &&
    isDescendantOf(targetNode, sourceNode.id, nodeIndex)
  ) {
    return {
      ...connection,
      source: sourceNode.id,
      sourceHandle: 'message:loop',
      target: targetNode.id,
      targetHandle: preferredMessageInputHandle(targetNode.type),
    };
  }

  if (
    targetNode.type === 'while_node' &&
    isWhileIngressHandle(connection.targetHandle) &&
    isDescendantOf(sourceNode, targetNode.id, nodeIndex) &&
    handleRole(sourceNode.type, connection.sourceHandle) !== 'source'
  ) {
    return {
      ...connection,
      source: targetNode.id,
      sourceHandle: 'message:loop',
      target: sourceNode.id,
      targetHandle: preferredMessageInputHandle(sourceNode.type),
    };
  }

  return connection;
};

export interface GraphState {
  nodes: Node[];
  edges: Edge[];
  version: string;
  name: string;
  validationErrors: string[];
  selectedNodeId: string | null;

  // React Flow Handlers
  onNodesChange: OnNodesChange;
  onEdgesChange: OnEdgesChange;
  onConnect: OnConnect;

  // Actions
  setSelectedNodeId: (nodeId: string | null) => void;
  setName: (name: string) => void;
  updateNodeConfig: (nodeId: string, config: any) => void;
  addNode: (type: string, position: { x: number, y: number }) => void;
  clearGraph: () => void;

  // Contract Serialization
  exportGraph: () => Graph;
  importGraph: (data: any) => boolean;
  validateGraph: () => boolean;
}

const getNodeWidth = (node: Node) => {
  const styleWidth = typeof node.style?.width === 'number' ? node.style.width : undefined;
  return styleWidth ?? node.width ?? (node.type === 'while_node' ? DEFAULT_WHILE_WIDTH : DEFAULT_NODE_WIDTH);
};

const getNodeHeight = (node: Node) => {
  const styleHeight = typeof node.style?.height === 'number' ? node.style.height : undefined;
  return styleHeight ?? node.height ?? (node.type === 'while_node' ? DEFAULT_WHILE_HEIGHT : DEFAULT_NODE_HEIGHT);
};

const persistedDimension = (value: number | null | undefined) => {
  return typeof value === 'number' ? value : undefined;
};

const buildNodeIndex = (nodes: Node[]) => new Map(nodes.map((node) => [node.id, node]));

const getAbsolutePosition = (node: Node, nodeIndex: Map<string, Node>) => {
  let x = node.position.x;
  let y = node.position.y;
  let current = node;
  const seen = new Set<string>([node.id]);

  while (current.parentNode) {
    const parent = nodeIndex.get(current.parentNode);
    if (!parent || seen.has(parent.id)) {
      break;
    }
    x += parent.position.x;
    y += parent.position.y;
    seen.add(parent.id);
    current = parent;
  }

  return { x, y };
};

const isDescendantOf = (node: Node, possibleAncestorID: string, nodeIndex: Map<string, Node>) => {
  let currentParentID = node.parentNode;
  const seen = new Set<string>();

  while (currentParentID) {
    if (currentParentID === possibleAncestorID) {
      return true;
    }
    if (seen.has(currentParentID)) {
      break;
    }
    seen.add(currentParentID);
    currentParentID = nodeIndex.get(currentParentID)?.parentNode;
  }

  return false;
};

const findContainingWhileContainer = (node: Node, absolutePosition: { x: number; y: number }, nodes: Node[]) => {
  const nodeIndex = buildNodeIndex(nodes);
  const candidates = nodes
    .filter((candidate) => candidate.type === 'while_node' && candidate.id !== node.id && !isDescendantOf(candidate, node.id, nodeIndex))
    .filter((candidate) => {
      const absolute = getAbsolutePosition(candidate, nodeIndex);
      const width = getNodeWidth(candidate);
      const height = getNodeHeight(candidate);
      return absolutePosition.x >= absolute.x &&
        absolutePosition.y >= absolute.y &&
        absolutePosition.x <= absolute.x + width &&
        absolutePosition.y <= absolute.y + height;
    })
    .sort((a, b) => (getNodeWidth(a) * getNodeHeight(a)) - (getNodeWidth(b) * getNodeHeight(b)));

  return candidates[0] ?? null;
};

const attachNodeToParent = (node: Node, parent: Node | null, absolutePosition: { x: number; y: number }, nodeIndex: Map<string, Node>) => {
  if (!parent) {
    return {
      ...node,
      position: absolutePosition,
      parentNode: undefined,
    };
  }

  const parentAbsolute = getAbsolutePosition(parent, nodeIndex);
  return {
    ...node,
    position: {
      x: absolutePosition.x - parentAbsolute.x,
      y: absolutePosition.y - parentAbsolute.y,
    },
    parentNode: parent.id,
  };
};

const reconcileContainerMembership = (nodes: Node[], changes: NodeChange[]) => {
  const changedNodeIDs = changes
    .filter((change): change is Extract<NodeChange, { type: 'position' }> => change.type === 'position')
    .filter((change) => change.dragging === false)
    .map((change) => change.id);

  if (changedNodeIDs.length === 0) {
    return nodes;
  }

  let nextNodes = nodes;
  for (const nodeID of changedNodeIDs) {
    const nodeIndex = buildNodeIndex(nextNodes);
    const node = nodeIndex.get(nodeID);
    if (!node || node.type === 'input_node') {
      continue;
    }

    const absolutePosition = getAbsolutePosition(node, nodeIndex);
    const container = findContainingWhileContainer(node, absolutePosition, nextNodes);
    const nextParentID = container?.id;
    if (node.parentNode === nextParentID) {
      continue;
    }

    nextNodes = nextNodes.map((candidate) => {
      if (candidate.id !== nodeID) {
        return candidate;
      }
      return attachNodeToParent(candidate, container, absolutePosition, nodeIndex);
    });
  }

  return nextNodes;
};

const toCanvasNode = (node: ContractNode): Node => {
  const style = node.width || node.height
    ? {
        ...(node.width ? { width: node.width } : {}),
        ...(node.height ? { height: node.height } : {}),
      }
    : node.type === 'while_node'
      ? { width: DEFAULT_WHILE_WIDTH, height: DEFAULT_WHILE_HEIGHT }
      : undefined;

  return {
    id: node.id,
    type: node.type,
    position: node.position,
    data: { config: node.config },
    parentNode: node.parent_id,
    style,
  };
};

// useGraphStore is the single source of truth for the canvas, serialized graph
// contract, and lightweight editor state such as selection and validation
// errors.
export const useGraphStore = create<GraphState>((set, get) => ({
  nodes: [],
  edges: [],
  version: "1.0",
  name: "New Workflow",
  validationErrors: [],
  selectedNodeId: null,

  onNodesChange: (changes: NodeChange[]) => {
    const nextNodes = reconcileContainerMembership(
      applyNodeChanges(changes, get().nodes),
      changes,
    );

    set({
      nodes: nextNodes,
    });
    // Track selection
    const selectionChange = changes.find(c => c.type === 'select');
    if (selectionChange && 'selected' in selectionChange) {
        if (selectionChange.selected) {
            set({ selectedNodeId: selectionChange.id });
        } else if (get().selectedNodeId === selectionChange.id) {
            set({ selectedNodeId: null });
        }
    }
  },

  onEdgesChange: (changes: EdgeChange[]) => {
    set({
      edges: applyEdgeChanges(changes, get().edges),
    });
  },

  onConnect: (connection: Connection) => {
    const directedConnection = normalizeConnectionDirection(connection, get().nodes);
    const canonicalConnection = normalizeWhileContainerConnection(directedConnection, get().nodes);
    // Saved graphs may still carry legacy handle ids such as "out_message".
    // Normalize them before validating compatibility or storing the edge so the
    // canvas always converges on one canonical port naming scheme.
    const sourceType = portFamily(canonicalConnection.sourceHandle);
    const targetType = portFamily(canonicalConnection.targetHandle);

    if (!sourceType || !targetType) {
        console.warn("Invalid connection: missing source or target handle", canonicalConnection);
        return;
    }

    if (sourceType !== targetType) {
        console.warn("Invalid connection: types must match", sourceType, targetType);
        return;
    }

    const normalizedConnection: Connection = {
      ...canonicalConnection,
      sourceHandle: canonicalConnection.sourceHandle || null,
      targetHandle: canonicalConnection.targetHandle || null,
    };

    const nodeTypes = new Map(get().nodes.map((node) => [node.id, node.type]));
    if (normalizedConnection.sourceHandle === 'message:loop') {
      normalizedConnection.targetHandle = preferredMessageInputHandle(nodeTypes.get(normalizedConnection.target || ''));
    }

    let nextEdges = get().edges;
    if (
      normalizedConnection.source &&
      nodeTypes.get(normalizedConnection.source) === 'while_node' &&
      normalizePort(normalizedConnection.sourceHandle) === 'message:loop'
    ) {
      nextEdges = nextEdges.filter(
        (edge) =>
          !(edge.source === normalizedConnection.source && normalizePort(edge.sourceHandle) === 'message:loop'),
      );
    }

    set({
      edges: addEdge({ ...normalizedConnection, type: 'workflow' }, nextEdges),
    });
  },

  setSelectedNodeId: (nodeId) => set({ selectedNodeId: nodeId }),

  setName: (name) => set({ name }),

  updateNodeConfig: (nodeId, config) => {
    set({
      nodes: get().nodes.map((node) =>
        node.id === nodeId ? { ...node, data: { ...node.data, config } } : node
      ),
    });
  },

  addNode: (type, position) => {
    const id = `${type}-${Date.now()}`;
    let config: any = {};
    let parentNode: string | undefined;
    let style: Node['style'];

    // New nodes start with contract-valid defaults so the canvas can emit a
    // valid graph as soon as possible after a drag-and-drop action.
    switch (type) {
      case 'input_node':
        config = { name: 'user_input', description: 'The user input' };
        break;
      case 'llm_node':
        config = { 
            name: 'llm_agent', 
            description: 'The Agent', 
            model: 'gemini-2.5-flash', 
            instruction: 'You are a helpful assistant.',
            response_mode: 'text',
            generate_content_config: { temperature: 0.7, max_output_tokens: 4096 }
        };
        break;
      case 'toolbox':
        config = { tools: [], mcp_servers: [], custom_functions: [] };
        break;
      case 'output_node':
        config = { name: 'final_output', output_key: 'result', format: 'message' };
        break;
      case 'if_else_node':
        config = { condition_language: 'CEL', condition: 'state.analyze.status == "pass"' };
        break;
      case 'while_node':
        config = { condition: 'state.analyze.status != "pass"', max_iterations: 3 };
        style = { width: DEFAULT_WHILE_WIDTH, height: DEFAULT_WHILE_HEIGHT };
        break;
    }

    if (type !== 'input_node') {
      const provisionalNode: Node = {
        id,
        type,
        position,
        data: { config },
      };
      const container = findContainingWhileContainer(provisionalNode, position, get().nodes);
      if (container) {
        const nodeIndex = buildNodeIndex(get().nodes);
        const parentAbsolute = getAbsolutePosition(container, nodeIndex);
        parentNode = container.id;
        position = {
          x: type === 'while_node'
            ? position.x - parentAbsolute.x
            : Math.max(position.x - parentAbsolute.x, CONTAINER_DROP_PADDING_X),
          y: type === 'while_node'
            ? position.y - parentAbsolute.y
            : Math.max(position.y - parentAbsolute.y, CONTAINER_DROP_PADDING_Y),
        };
      }
    }

    const newNode: Node = {
      id,
      type,
      position,
      data: { config },
      parentNode,
      style,
    };

    set({
      nodes: [...get().nodes, newNode],
    });
  },

  clearGraph: () => set({ nodes: [], edges: [], name: 'New Workflow', validationErrors: [], selectedNodeId: null }),

  exportGraph: () => {
    const { nodes, edges, version, name } = get();
    
    // React Flow node metadata is richer than the persisted contract, so export
    // strips the canvas representation down to the JSON document shared with
    // the backend.
    const contractNodes: ContractNode[] = nodes.map(node => ({
      id: node.id,
      type: node.type as any,
      position: node.position,
      parent_id: node.parentNode || undefined,
      width: persistedDimension(typeof node.style?.width === 'number' ? node.style.width : node.width),
      height: persistedDimension(typeof node.style?.height === 'number' ? node.style.height : node.height),
      config: node.data.config,
    }));

    const contractEdges: ContractEdge[] = edges.map(edge => ({
      id: edge.id,
      source: edge.source,
      source_port: normalizePort(edge.sourceHandle) || "",
      target: edge.target,
      target_port: normalizePort(edge.targetHandle) || "",
      data_type: portFamily(edge.sourceHandle),
      edge_kind: "data_flow",
    }));

    return {
      version,
      name,
      nodes: contractNodes,
      edges: contractEdges,
    };
  },

  importGraph: (data: any) => {
    try {
      const validated = GraphSchema.parse(data);
      
      // Import performs the inverse of exportGraph: take a contract document and
      // inflate it back into the subset of React Flow fields the editor needs.
      const rfNodes: Node[] = validated.nodes.map(toCanvasNode);

      const rfEdges: Edge[] = validated.edges.map(edge => ({
        id: edge.id,
        type: 'workflow',
        source: edge.source,
        sourceHandle: toCanvasSourcePort(rfNodes.find((node) => node.id === edge.source)?.type, edge.source_port),
        target: edge.target,
        targetHandle: toCanvasTargetPort(rfNodes.find((node) => node.id === edge.target)?.type, edge.target_port),
      }));

      set({
        nodes: rfNodes,
        edges: rfEdges,
        version: validated.version,
        name: validated.name,
        validationErrors: [],
        selectedNodeId: null,
      });
      return true;
    } catch (err: any) {
      console.error("Import failed:", err);
      if (err.errors) {
        set({ validationErrors: err.errors.map((e: any) => `${e.path.join('.')}: ${e.message}`) });
      }
      return false;
    }
  },

  validateGraph: () => {
    // Frontend validation catches schema drift early before a graph is sent to
    // the backend, while the backend still remains the source of truth for
    // runtime-supported behavior.
    const graph = get().exportGraph();
    const result = GraphSchema.safeParse(graph);
    if (!result.success) {
      const errorStrings = result.error?.errors.map(e => `${e.path.join('.')}: ${e.message}`) || ["Unknown validation error"];
      set({ validationErrors: errorStrings });
      return false;
    } else {
      set({ validationErrors: [] });
      return true;
    }
  },
}));
