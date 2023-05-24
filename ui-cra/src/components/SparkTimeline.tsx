/* eslint-disable id-length,no-mixed-operators */

import { scaleLinear } from 'd3-scale';
import { utcHour } from 'd3-time';
import { entries, groupBy, range } from 'lodash';
import moment from 'moment';
import { FC, Key } from 'react';
import { useMeasure } from 'react-use';
import styled from 'styled-components';

export const COMPACT_LOCALE_KEY = 'compact-time-ranges';

// FIXME: move this somewhere else?
(() => {
  // When you register a new locale moment.js changes the global to
  // the new entry. So save the current locale and then set it back.
  const defaultLocale = moment.locale();
  moment.locale(COMPACT_LOCALE_KEY, {
    relativeTime: {
      d: '1d',
      dd: '%dd',
      future: 'in %s',
      h: '1h',
      hh: '%dh',
      m: '1m',
      M: '1mo',
      mm: '%dm',
      MM: '%dmo',
      past: '%s ago',
      s: '%ds',
      y: '1y',
      yy: '%dy',
    },
  });
  moment.locale(defaultLocale);
})();

const Container = styled.div`
  width: 100%;
  height: 100%;

  svg {
    overflow: visible;
  }

  font-family: sans-serif;
  font-size: 11px;
`;

const Tick = styled.line`
  stroke: ${props => props.theme.colors.neutral30};
`;

const MajorTick = styled.line`
  stroke: ${props => props.theme.colors.neutral30};
`;

const statusColor = (props: { status: string }) => {
  const colorIndex: { [key: string]: string } = {
    success: '#27AE60',
    fail: '#BC3B1D',
  };
  return colorIndex[props.status] || '#00b3ec';
};

const Point = styled.circle<{ status: string }>`
  stroke: ${statusColor};
  fill: ${props => props.theme.colors.white};
  stroke-width: 2px;
`;

const PointLayer = styled.g<{ status: string }>`
  cursor: pointer;
  &:hover ${Point} {
    fill: ${statusColor};
  }
`;

const HeadText = styled.text`
  && {
    color: ${props => props.theme.colors.neutral30};
    font-size: 9px;
  }
`;

const SvgCanvas = styled.svg<{ axisOnHover: boolean }>`
  text {
    opacity: ${props => (props.axisOnHover ? 0 : 1)};
    cursor: default;
    transition: opacity 0.2s linear;
    color: ${props => props.theme.colors.neutral40};
  }
  &:hover {
    ${MajorTick} {
      stroke: ${props => props.theme.colors.neutral30};
      stroke-width: 2px;
    }
    text {
      opacity: 1;
    }
  }
`;

type CommitRenderer = (commit: JSX.Element, key: Key, data: any) => JSX.Element;

interface PointType {
  ts: Date;
  status: string;
}

interface SparkTimelineProps {
  axisOnHover: boolean;
  data: PointType[];
  showHeadLabel: boolean;
  renderCommit?: CommitRenderer;
}

const hoursAgo = (now: Date, date: Date): number =>
  (utcHour.floor(now).getTime() - utcHour.floor(date).getTime()) /
  1000 /
  60 /
  60;

export const SparkTimeline: FC<SparkTimelineProps> = ({ data, ...props }) => {
  const now = new Date();
  const groupedData = entries(groupBy(data, d => hoursAgo(now, d.ts))).map(
    ([k, v]) => ({
      index: parseInt(k, 10),
      data: v,
    }),
  );

  return <SparkTimelineInner data={groupedData} {...props} />;
};

interface GroupedPoints {
  index: number;
  data: PointType[];
}

interface SparkTimelineInnerProps {
  axisOnHover: boolean;
  data: GroupedPoints[];
  showHeadLabel: boolean;
  renderCommit?: CommitRenderer;
}

const SparkTimelineInner: FC<SparkTimelineInnerProps> = ({
  axisOnHover,
  data,
  showHeadLabel,
  renderCommit,
}) => {
  const [ref, { width }] = useMeasure();

  const { true: outOfRange, false: inRange } = groupBy(
    data,
    d => d.index >= 24,
  );
  const points = inRange || [];
  const older = outOfRange || [];

  const commitRadius = 4;
  const commitStrokeWidth = 2;
  const height = (commitRadius + commitStrokeWidth) * 2 + 14 * 2;
  const padding = commitRadius;
  const olderCommitPadding = 24;
  const innerWidth = width - padding * 2;
  const x = scaleLinear()
    .domain([23, 0])
    .range([0, innerWidth - olderCommitPadding]);

  const oldCommit = older[0];
  const lastCommit = points[0];
  const headOffset = lastCommit ? x(lastCommit.index) + olderCommitPadding : 0;
  const noCommits = !oldCommit && !lastCommit;
  const oldCommitString =
    oldCommit &&
    moment(new Date(oldCommit.data[0].ts))
      .utc()
      .locale(COMPACT_LOCALE_KEY)
      .fromNow(true);

  const innerCommitRender = renderCommit ?? (el => el);

  // FIXME: avoid any somehow
  return (
    <Container ref={ref as any}>
      <SvgCanvas axisOnHover={axisOnHover} width={width} height={height}>
        <g transform={`translate(${padding}, 22)`}>
          <text
            fill="currentColor"
            x={olderCommitPadding}
            y="10"
            textAnchor="middle"
            dy="0.71em"
          >
            24h
          </text>
          <text
            fill="currentColor"
            x={innerWidth}
            y="10"
            textAnchor="middle"
            dy="0.71em"
          >
            now
          </text>
          {!noCommits && showHeadLabel && (
            <HeadText
              fill="currentColor"
              x={headOffset}
              textAnchor="middle"
              y="-18"
              dy="0.71em"
            >
              HEAD
            </HeadText>
          )}
          {oldCommit && (
            <>
              <text fill="currentColor" y="10" textAnchor="middle" dy="0.71em">
                {oldCommitString}
              </text>
              {innerCommitRender(
                <PointLayer status={oldCommit.data[0].status}>
                  <Point status={oldCommit.data[0].status} r={commitRadius} />
                </PointLayer>,
                'old-commit',
                oldCommit.data,
              )}
            </>
          )}
          <g transform={`translate(${olderCommitPadding}, 0)`}>
            {range(24).map(ts => (
              <g transform={`translate(${x(ts)}, 0)`} key={ts}>
                {ts === 0 || ts === 23 ? (
                  <MajorTick y1="-3" y2="3" />
                ) : (
                  <Tick y1="-3" y2="3" />
                )}
              </g>
            ))}
            {points.map(({ index, data }) =>
              innerCommitRender(
                <PointLayer
                  status={data[0].status}
                  transform={`translate(${x(index)}, 0)`}
                  key={index}
                >
                  {data.length > 1 && (
                    <Point cy={3} status={data[0].status} r={commitRadius} />
                  )}
                  <Point status={data[0].status} r={commitRadius} />
                </PointLayer>,
                index,
                data,
              ),
            )}
          </g>
        </g>
      </SvgCanvas>
    </Container>
  );
};

/* eslint-enable id-length,no-mixed-operators */
