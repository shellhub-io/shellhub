interface ArrowMarkerProps {
  id: string;
  fill: string;
  markerWidth?: number;
  markerHeight?: number;
  refX?: number;
  refY?: number;
}

/**
 * An SVG arrowhead, for referencing from a path's marker-end. It renders a <marker>, so it only
 * has meaning inside an <svg> <defs>, and id is what a path points at — it must be unique in
 * the document.
 */
export function ArrowMarker({
  id,
  fill,
  markerWidth = 8,
  markerHeight = 6,
  refX = 8,
  refY = 3,
}: ArrowMarkerProps) {
  return (
    <marker
      id={id}
      markerWidth={markerWidth}
      markerHeight={markerHeight}
      refX={refX}
      refY={refY}
      orient="auto"
    >
      <path
        d={`M0,0 L${markerWidth},${markerHeight / 2} L0,${markerHeight}`}
        fill={fill}
      />
    </marker>
  );
}
