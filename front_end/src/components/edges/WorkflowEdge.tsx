import React, { memo } from 'react';
import { BaseEdge, EdgeProps, Position, getSmoothStepPath, useStore } from 'reactflow';

const WHILE_ENTER_RATIO = 0.46;
const WHILE_REPEAT_RATIO = 0.72;
const WHILE_HANDLE_CENTER_OFFSET_X = 5;

const edgeStyle = (style?: React.CSSProperties): React.CSSProperties => ({
  stroke: '#94a3b8',
  strokeWidth: 2,
  strokeLinecap: 'round',
  ...style,
});

const whileAnchor = (
  node: { positionAbsolute?: { x: number; y: number }; width?: number | null; height?: number | null } | undefined,
  ratio: number,
) => {
  if (!node?.positionAbsolute || typeof node.height !== 'number') {
    return null;
  }

  return {
    x: node.positionAbsolute.x - WHILE_HANDLE_CENTER_OFFSET_X,
    y: node.positionAbsolute.y + (node.height * ratio),
  };
};

const smoothPath = (
  sourceX: number,
  sourceY: number,
  sourcePosition: Position,
  targetX: number,
  targetY: number,
  targetPosition: Position,
) =>
  getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
    borderRadius: 22,
    offset: 42,
  })[0];

export const WorkflowEdge = memo((props: EdgeProps) => {
  const sourceNode = useStore((state) => state.nodeInternals.get(props.source));
  const targetNode = useStore((state) => state.nodeInternals.get(props.target));
  const pathStyle = edgeStyle(props.style);

  const isWhileLoopEdge = sourceNode?.type === 'while_node' && props.sourceHandleId === 'message:loop';
  if (!isWhileLoopEdge) {
    const path = smoothPath(
      props.sourceX,
      props.sourceY,
      props.sourcePosition,
      props.targetX,
      props.targetY,
      props.targetPosition,
    );

    return (
      <BaseEdge
        path={path}
        markerStart={props.markerStart}
        markerEnd={props.markerEnd}
        interactionWidth={props.interactionWidth}
        style={pathStyle}
      />
    );
  }

  const enterAnchor = whileAnchor(sourceNode, WHILE_ENTER_RATIO);
  const repeatAnchor = whileAnchor(sourceNode, WHILE_REPEAT_RATIO);
  const targetPosition = targetNode?.targetPosition ?? props.targetPosition ?? Position.Left;

  const enterPath = smoothPath(
    enterAnchor?.x ?? props.sourceX,
    enterAnchor?.y ?? props.sourceY,
    Position.Left,
    props.targetX,
    props.targetY,
    targetPosition,
  );
  const repeatPath = smoothPath(
    repeatAnchor?.x ?? props.sourceX,
    repeatAnchor?.y ?? props.sourceY,
    Position.Left,
    props.targetX,
    props.targetY,
    targetPosition,
  );

  return (
    <>
      <BaseEdge
        path={enterPath}
        markerStart={props.markerStart}
        markerEnd={props.markerEnd}
        interactionWidth={props.interactionWidth}
        style={pathStyle}
      />
      <BaseEdge
        path={repeatPath}
        markerEnd={props.markerEnd}
        style={{ ...pathStyle, opacity: 0.92 }}
      />
    </>
  );
});

WorkflowEdge.displayName = 'WorkflowEdge';
