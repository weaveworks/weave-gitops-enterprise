import { filterConfig, theme, FilterableTable } from '@weaveworks/weave-gitops';
import { ThemeProvider } from 'styled-components';
import { usePolicyStyle } from '../../../Policies/PolicyStyles';
import { TableWrapper } from '../../CanaryStyles';
import {CanaryMetric} from "@weaveworks/progressive-delivery/api/prog/types.pb";
import {GitopsClusterEnriched} from "../../../../types/custom";

export const CanaryMetricsTable = ({ metrics }: { metrics: CanaryMetric[] }) => {
    const classes = usePolicyStyle();

    const initialFilterState = {
        // ...filterConfig(events, 'component'),
    };

    return (
        <div className={classes.root}>
            <ThemeProvider theme={theme}>
                {metrics.length > 0 ? (
                    <TableWrapper id="canary-analysis-metrics">
                        <FilterableTable
                            key={metrics?.length}
                            filters={initialFilterState}
                            rows={metrics}
                            fields={[
                                {
                                    label: 'Name',
                                    value: 'name',
                                    // textSearchable: true,
                                },
                                // {
                                //     label: 'Namespace',
                                //     value: (c: CanaryMetric) =>
                                //         (c.namespace) || "no-namespace",
                                //     // textSearchable: true,
                                // },
                                {
                                    label: 'Threshold Min',
                                    value: (c: CanaryMetric) =>
                                        "" + (c.thresholdRange?.min),
                                },
                                {
                                    label: 'Threshold Max',
                                    value: (c: CanaryMetric) =>
                                        "" + (c.thresholdRange?.max),
                                },
                                {
                                    label: 'Interval',
                                    value: 'interval',
                                }
                            ]}
                        />
                    </TableWrapper>
                ) : (
                    <p>No metrics to display</p>
                )}
            </ThemeProvider>
        </div>
    );
};
