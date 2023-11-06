import { Switch, Tooltip } from 'antd';
import classNames from 'classnames';
import { useContext, useMemo, useState } from 'react';
import { generatePath, useNavigate } from 'react-router-dom';

import { paths } from '@ui/config/paths';
import { ColorContext } from '@ui/context/colors';
import { Stage } from '@ui/gen/v1alpha1/types_pb';
import { useLocalStorage } from '@ui/utils/use-local-storage';

interface StagePixelStyle {
  opacity: number;
  backgroundColor: string;
  border?: string;
}

type StageStyleMap = { [key: string]: StagePixelStyle };

const ChartVersionRow = ({
  version,
  stages,
  stylesByStage,
  projectName,
  showHistory
}: {
  version: string;
  stages: Stage[];
  stylesByStage: StageStyleMap;
  projectName: string;
  showHistory: boolean;
}) => {
  const navigate = useNavigate();
  return (
    <div className='flex items-center mb-2'>
      <Tooltip title={version}>
        <div className='mr-4 font-mono text-sm text-right w-20 truncate'>{version}</div>
      </Tooltip>
      {stages.map((stage) => {
        let curStyles: StagePixelStyle | null = stylesByStage[stage.metadata?.name || ''];
        if (curStyles) {
          if (!showHistory && curStyles.opacity < 1) {
            curStyles = null;
          } else if (showHistory && curStyles.opacity == 1) {
            curStyles = {
              ...curStyles,
              border: '3px solid rgba(255,255,255,0.3)'
            };
          }
        }
        return (
          <Tooltip key={stage.metadata?.name} title={stage.metadata?.name}>
            <div
              className={classNames('mr-2 bg-zinc-600 ', {
                'cursor-pointer': !!curStyles
              })}
              style={{
                borderRadius: '5px',
                border: '3px solid transparent',
                height: '30px',
                width: '30px',
                ...curStyles
              }}
              onClick={() =>
                navigate(
                  generatePath(paths.stage, {
                    name: projectName,
                    stageName: stage.metadata?.name
                  })
                )
              }
            />
          </Tooltip>
        );
      })}
    </div>
  );
};

export const Charts = ({ projectName, stages }: { projectName: string; stages: Stage[] }) => {
  const colors = useContext(ColorContext);
  const charts = useMemo(() => {
    const charts = new Map<string, Map<string, StageStyleMap>>();
    stages.forEach((stage) => {
      const len = stage.status?.history?.length || 0;
      stage.status?.history?.forEach((freight, i) => {
        freight.charts?.forEach((chart) => {
          let registry = charts.get(chart.registryUrl);
          if (!registry) {
            registry = new Map<string, StageStyleMap>();
            charts.set(chart.registryUrl, registry);
          }
          let stages = registry.get(chart.version);
          if (!stages) {
            stages = {} as StageStyleMap;
            registry.set(chart.version, stages);
          }
          stages[stage.metadata?.name as string] = {
            opacity: 1 - i / len,
            backgroundColor: colors[stage.metadata?.uid as string]
          };
        });
      });

      stage.status?.currentFreight?.charts?.forEach((chart) => {
        let registry = charts.get(chart.registryUrl);
        if (!registry) {
          registry = new Map<string, StageStyleMap>();
          charts.set(chart.registryUrl, registry);
        }
        let stages = registry.get(chart.version);
        if (!stages) {
          stages = {} as StageStyleMap;
          registry.set(chart.version, stages);
        }
        stages[stage.metadata?.name as string] = {
          opacity: 1,
          backgroundColor: colors[stage.metadata?.uid as string]
        };
      });
    });
    return charts;
  }, [stages]);

  const [chartUrl, setChartUrl] = useState(charts.keys().next().value as string);
  const chart = chartUrl && charts.get(chartUrl);
  const [showHistory, setShowHistory] = useLocalStorage(`${projectName}-show-history`, false);

  return (
    <>
      {chart ? (
        <>
          <div className='mb-4 flex items-center'>
            <Switch onChange={(val) => setShowHistory(val)} checked={showHistory} />
            <div className='ml-2 font-semibold'>SHOW HISTORY</div>
          </div>
          <div className='mb-8'>
            <Select
              value={chartUrl}
              onChange={(value) => setChartUrl(value as string)}
              options={Array.from(charts.keys()).map((chart) => ({
                label: chart.name,
                value: chart
              }))}
            />
          </div>
          {Array.from(chart.entries())
            .sort((a, b) => b[0].localeCompare(a[0], undefined, { numeric: true }))
            .map(([version, versionStages]) => (
              <ChartVersionRow
                key={version}
                projectName={projectName}
                version={version}
                stages={stages}
                stylesByStage={versionStages}
                showHistory={showHistory}
              />
            ))}
        </>
      ) : (
        <p>No charts available</p>
      )}
    </>
  );
};

const Select = ({
  value,
  onChange,
  options
}: {
  value: string;
  onChange: (value: string) => void;
  options: { label?: string; value: string }[];
}) => (
  <select
    className='block border-none w-full text-gray appearance-none p-2 bg-zinc-700 focus:outline-none focus:ring-2 focus:ring-blue-400'
    value={value}
    onChange={(e) => onChange(e.target.value)}
  >
    {options.map((option) => (
      <option value={option.value} key={option.label}>
        {option.label}
      </option>
    ))}
  </select>
);
