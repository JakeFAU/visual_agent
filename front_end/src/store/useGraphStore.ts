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
  updateNodeConfig: (nodeId: string, config: any) => void;
  addNode: (type: string, position: { x: number, y: number }) => void;
  clearGraph: () => void;

  // Contract Serialization
  exportGraph: () => Graph;
  importGraph: (data: any) => void;
  validateGraph: () => boolean;
}

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
    set({
      nodes: applyNodeChanges(changes, get().nodes),
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
    // Ports encode a logical data type prefix such as "message" or "logic".
    // For branch handles the suffix after ":" identifies the branch name, so
    // only the prefix participates in compatibility checks.
    const sourceType = connection.sourceHandle?.split(':')[0];
    const targetType = connection.targetHandle?.split(':')[0];

    if (sourceType !== targetType) {
        console.warn("Invalid connection: types must match", sourceType, targetType);
        return;
    }
    set({
      edges: addEdge(connection, get().edges),
    });
  },

  setSelectedNodeId: (nodeId) => set({ selectedNodeId: nodeId }),

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
        config = { condition_language: 'CEL', condition: 'state.category == "billing"' };
        break;
    }

    const newNode: Node = {
      id,
      type,
      position,
      data: { config },
    };

    set({
      nodes: [...get().nodes, newNode],
    });
  },

  clearGraph: () => set({ nodes: [], edges: [], name: 'New Workflow', validationErrors: [] }),

  exportGraph: () => {
    const { nodes, edges, version, name } = get();
    
    // React Flow node metadata is richer than the persisted contract, so export
    // strips the canvas representation down to the JSON document shared with
    // the backend.
    const contractNodes: ContractNode[] = nodes.map(node => ({
      id: node.id,
      type: node.type as any,
      position: node.position,
      config: node.data.config,
    }));

    const contractEdges: ContractEdge[] = edges.map(edge => ({
      id: edge.id,
      source: edge.source,
      source_port: edge.sourceHandle || "",
      target: edge.target,
      target_port: edge.targetHandle || "",
      data_type: (edge.sourceHandle || "").split(':')[0],
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
      const rfNodes: Node[] = validated.nodes.map(node => ({
        id: node.id,
        type: node.type,
        position: node.position,
        data: { config: node.config },
      }));

      const rfEdges: Edge[] = validated.edges.map(edge => ({
        id: edge.id,
        source: edge.source,
        sourceHandle: edge.source_port,
        target: edge.target,
        targetHandle: edge.target_port,
      }));

      set({
        nodes: rfNodes,
        edges: rfEdges,
        version: validated.version,
        name: validated.name,
        validationErrors: [],
      });
    } catch (err: any) {
      console.error("Import failed:", err);
      if (err.errors) {
        set({ validationErrors: err.errors.map((e: any) => `${e.path.join('.')}: ${e.message}`) });
      }
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
