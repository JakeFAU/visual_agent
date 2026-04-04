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

  // Selection
  setSelectedNodeId: (nodeId: string | null) => void;

  // Node Mutation
  updateNodeConfig: (nodeId: string, config: any) => void;

  // Contract Serialization
  exportGraph: () => Graph;
  importGraph: (json: string) => void;
  validateGraph: () => void;
}

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
    // Validate types before connecting
    if (connection.sourceHandle !== connection.targetHandle) {
        console.warn("Invalid connection: types must match", connection.sourceHandle, connection.targetHandle);
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

  exportGraph: () => {
    const { nodes, edges, version, name } = get();
    
    const contractNodes: ContractNode[] = nodes.map(node => ({
      id: node.id,
      type: node.type as any, // Validated by Zod later
      position: node.position,
      config: node.data.config,
    }));

    const contractEdges: ContractEdge[] = edges.map(edge => ({
      id: edge.id,
      source: edge.source,
      source_port: edge.sourceHandle || "",
      target: edge.target,
      target_port: edge.targetHandle || "",
      data_type: edge.sourceHandle || "", // Assuming handle ID matches data_type
      edge_kind: "data_flow",
    }));

    return {
      version,
      name,
      nodes: contractNodes,
      edges: contractEdges,
    };
  },

  importGraph: (json: string) => {
    try {
      const data = JSON.parse(json);
      const validated = GraphSchema.parse(data);
      
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
    const graph = get().exportGraph();
    const result = GraphSchema.safeParse(graph);
    if (!result.success) {
      const errorStrings = result.error?.errors.map(e => `${e.path.join('.')}: ${e.message}`) || ["Unknown validation error"];
      set({ validationErrors: errorStrings });
    } else {
      set({ validationErrors: [] });
    }
  },
}));
