import {
  Checkbox,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableFooter,
  TableHead,
  TableRow,
  TableSortLabel,
  Theme,
} from "@material-ui/core";
import { createStyles, makeStyles } from "@material-ui/styles";
import { range } from "lodash";
import React, { FC, useEffect } from "react";
import { Cluster } from "../../../types/kubernetes";
import { Pagination } from "../../Pagination";
import { ColumnHeaderTooltip, SkeletonRow } from "../../Shared";
import ClusterRow from "./Row";
import { muiTheme } from "../../../muiTheme";
import { ThemeProvider, createMuiTheme } from "@material-ui/core/styles";
import { Shadows } from "@material-ui/core/styles/shadows";
import useClusters from "../../../contexts/Clusters";
import useNotifications from "../../../contexts/Notifications";
import { useHistory } from "react-router-dom";
import { Loader } from "../../Loader";

const localMuiTheme = createMuiTheme({
  ...muiTheme,
  shadows: Array(25).fill("none") as Shadows,
});

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    nameHeaderCell: {
      paddingLeft: theme.spacing(4),
    },
    paper: {
      marginBottom: 10,
      marginTop: 10,
      overflowX: "auto",
      width: "100%",
    },
    root: {
      width: "100%",
    },
    disabled: {
      opacity: 0.5,
    },
    table: {
      whiteSpace: "nowrap",
    },
    tableHead: {
      borderBottom: "1px solid #d8d8d8",
    },
    noMaxWidth: {
      maxWidth: "none",
    },
    tablePagination: {
      height: "80px",
    },
  })
);

interface Props {
  filteredClusters: Cluster[] | null;
  count: number | null;
  disabled?: boolean;
  onEdit: (cluster: Cluster) => void;
  onSortChange: (order: string) => void;
  onSelectPageParams: (page: number, perPage: number) => void;
  order: string;
  orderBy: string;
}

export const ClustersTable: FC<Props> = ({
  filteredClusters,
  count,
  disabled,
  onEdit,
  onSortChange,
  onSelectPageParams,
  order,
  orderBy,
}) => {
  const classes = useStyles();
  const history = useHistory();
  const { selectedClusters, setSelectedClusters, creatingPR, loading } =
    useClusters();
  const { notifications } = useNotifications();
  const numSelected = selectedClusters.length;
  const isSelected = (name: string) => selectedClusters.indexOf(name) !== -1;
  const rowCount = filteredClusters?.length || 0;

  const showSkeleton = !filteredClusters;
  const skeletonRows = range(0, 1).map((id, index) => (
    <SkeletonRow index={index} key={id} />
  ));

  const handleSelectAllClick = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      const newSelected = filteredClusters?.map((cluster) => cluster.name);
      setSelectedClusters(newSelected || []);
      return;
    }
    setSelectedClusters([]);
  };

  const handleClick = (event: React.MouseEvent<unknown>, name: string) => {
    const selectedIndex = selectedClusters.indexOf(name);
    let newSelected: string[] = [];

    if (selectedIndex === -1) {
      newSelected = newSelected.concat(selectedClusters, name);
    } else if (selectedIndex === 0) {
      newSelected = newSelected.concat(selectedClusters.slice(1));
    } else if (selectedIndex === selectedClusters.length - 1) {
      newSelected = newSelected.concat(selectedClusters.slice(0, -1));
    } else if (selectedIndex > 0) {
      newSelected = newSelected.concat(
        selectedClusters.slice(0, selectedIndex),
        selectedClusters.slice(selectedIndex + 1)
      );
    }
    setSelectedClusters(newSelected);
  };

  useEffect(() => {
    return history.listen(() => {
      setSelectedClusters([]);
    });
  }, [notifications, history, setSelectedClusters]);

  return (
    <div
      className={`${classes.root} ${disabled ? classes.disabled : ""}`}
      id="clusters-list"
    >
      <ThemeProvider theme={localMuiTheme}>
        <Paper className={classes.paper}>
          {creatingPR || (loading && filteredClusters?.length === 0) ? (
            <Loader />
          ) : (
            <Table className={classes.table} size="small">
              {filteredClusters?.length === 0 ? (
                <caption>No clusters configured</caption>
              ) : null}
              <TableHead className={classes.tableHead}>
                <TableRow>
                  <TableCell padding="checkbox">
                    <Checkbox
                      indeterminate={numSelected > 0 && numSelected < rowCount}
                      checked={rowCount > 0 && numSelected === rowCount}
                      onChange={handleSelectAllClick}
                      inputProps={{ "aria-label": "select all clusters" }}
                      style={{
                        color: "#00B3EC",
                      }}
                    />
                  </TableCell>
                  <TableCell className={classes.nameHeaderCell} align="left">
                    <TableSortLabel
                      disabled={disabled}
                      active={orderBy === "Name"}
                      direction={
                        orderBy === "Name" ? (order as "asc" | "desc") : "asc"
                      }
                      onClick={() => onSortChange("Name")}
                    >
                      <ColumnHeaderTooltip title="Name configured in management UI">
                        <span>Name</span>
                      </ColumnHeaderTooltip>
                    </TableSortLabel>
                  </TableCell>
                  <TableCell />
                  <TableCell align="left">
                    <TableSortLabel
                      disabled={disabled}
                      active={orderBy === "ClusterStatus"}
                      direction={
                        orderBy === "ClusterStatus"
                          ? (order as "asc" | "desc")
                          : "asc"
                      }
                      onClick={() => onSortChange("ClusterStatus")}
                    >
                      <ColumnHeaderTooltip
                        title={
                          <span>
                            Shows the status of your clusters based on Agent
                            connection and Alertmanager alerts
                          </span>
                        }
                      >
                        <span>Status</span>
                      </ColumnHeaderTooltip>
                    </TableSortLabel>
                  </TableCell>
                  <TableCell align="left">
                    <ColumnHeaderTooltip title="Last commit to the cluster's git repository">
                      <span>Latest git activity</span>
                    </ColumnHeaderTooltip>
                  </TableCell>
                  <TableCell align="left">
                    <ColumnHeaderTooltip
                      classes={{ tooltip: classes.noMaxWidth }}
                      title="Kubernetes version ( [control plane nodes] | worker nodes)"
                    >
                      <span>Version ( Nodes )</span>
                    </ColumnHeaderTooltip>
                  </TableCell>
                  <TableCell align="left">
                    <ColumnHeaderTooltip title="Team Workspaces in the cluster">
                      <span>Team Workspaces</span>
                    </ColumnHeaderTooltip>
                  </TableCell>
                  <TableCell />
                  <TableCell />
                </TableRow>
              </TableHead>
              <TableBody>
                {showSkeleton && skeletonRows}
                {!showSkeleton &&
                  filteredClusters?.map((cluster: Cluster, index: number) => {
                    const isItemSelected = isSelected(cluster.name);
                    return (
                      <ClusterRow
                        key={cluster.name}
                        index={index}
                        cluster={cluster}
                        aria-checked={isItemSelected}
                        onCheckboxClick={(event) =>
                          handleClick(event, cluster.name)
                        }
                        onEdit={onEdit}
                        selected={isItemSelected}
                      />
                    );
                  })}
              </TableBody>
              <TableFooter>
                {filteredClusters?.length === 0 ? null : (
                  <TableRow>
                    <Pagination
                      className={classes.tablePagination}
                      count={count}
                      onSelectPageParams={onSelectPageParams}
                    />
                  </TableRow>
                )}
              </TableFooter>
            </Table>
          )}
        </Paper>
      </ThemeProvider>
    </div>
  );
};
