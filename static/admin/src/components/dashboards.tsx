
import * as React from 'react';
import {Link} from 'react-router-dom';
import {httpGet, httpDelete} from '../utils';

class componentState {
    rows: Array<any>;
    widgets: Array<Array<any>>;
    colWidth: number;
    dashboardID: number;
    dashboard: any;
}

class DashboardManagerComponent extends React.Component<any, componentState> {
    constructor(props: any) {
        super(props);
        this.state = {rows: [], widgets: [], colWidth: 0, dashboardID: 0, dashboard: null};
    }
    loadData() {
        httpGet('/api/v1/dashboards').then((response: any) => {
            this.setState({rows: (response.data.result ? response.data.result: [])});
        });
    }
    componentDidMount() {
        this.loadData();
    }
    showDashboardConfiguration(row: any) {

        var grid: Array<any> = [];

        for (let i = 0; i < row.height; i++) {
            const column = [];
            for (let h = 0; h < row.width; h++) {
                column.push({});
            }
            grid.push(column);
        }

        httpGet(`/api/v1/dashboards/${row.id}/widgets`).then((response) => {

            let widgets = (response.data.result || []);

            for (let w of widgets) {
                if (!grid[w.posY][w.posX]) {
                    console.log("no more columns to place the widget");
                    continue
                }
                grid[w.posY][w.posX] = w;
                grid[w.posY].splice(w.posX + 1, w.span - 1);
            }

            this.setState({
                dashboardID: row.id,
                dashboard: row,
                colWidth: 12 / row.width,
                widgets: grid,
            });
        })

    }
    deleteDashboard(dashboard: any) {
        httpDelete(`/api/v1/dashboards/${dashboard.id}`).then(() => {
            this.setState({widgets: []})
            this.loadData();
        })
    }
    removeWidget(widget: any) {
        httpDelete(`/api/v1/dashboards/${this.state.dashboardID}/widgets/${widget.id}`).then(() => {
            this.setState({widgets: []})
            this.showDashboardConfiguration(this.state.dashboard);
        })
    }
    render() {
    return (
        <React.Fragment>
            <div className="row">
                <div className="col-12">
                    <Link to="/dashboards/create" className="btn btn-primary">Add dashboard</Link>
                </div>
            </div>
            <div className="row">
                <div className="col-3">
                    <div className="card">
                        <ul className="list-group">
                            {this.state.rows.map((row) => {
                                return (
                                    <li className="list-group-item">
                                        <span onClick={() => this.showDashboardConfiguration(row)}>{row.name}</span>
                                        <button className="btn btn-danger float-right" onClick={(e) => this.deleteDashboard(row)}>
                                            Delete
                                        </button>
                                    </li>
                                )
                            })}
                        </ul>
                    </div>
                </div>
                <div className="col-9">
                    <div className="container-fluid">
                        {this.state.widgets.map((row, columnIndex) => {
                            return (
                            <div className="row">
                                {
                                    row.map((w, rowIndex) => {
                                        const maxSpan = ((12 - this.state.colWidth * (w.span || 1)) / this.state.colWidth) + 1;
                                        return (w.id ? (
                                            <div className={"col-" + (this.state.colWidth * w.span)}>
                                                <div className="card card-outline card-primary">
                                                    <div className="card-header">
                                                        {w.title}
                                                    </div>
                                                    <div className="card-body">
                                                        <button onClick={(e) => this.removeWidget(w)} className="btn btn-warning">
                                                            Remove
                                                        </button>
                                                    </div>
                                                </div>
                                            </div>
                                        ): (
                                            <div className={"col-" + this.state.colWidth}>
                                                <div className="card card-outline card-primary">
                                                    <div className="card-header">
                                                        <Link className="btn btn-primary" to={`/dashboards/widgets/${this.state.dashboardID}/add/${rowIndex}/${columnIndex}/${maxSpan}`}>
                                                            Add
                                                        </Link>
                                                    </div>
                                                    <div className="card-body text-center">
                                                        <i className="ion ion-md-add"></i>
                                                    </div>
                                                </div>
                                            </div>
                                        ) 
                                        )
                                    })
                                }     
                            </div>
                            )
                        })}
                    </div>
                </div>
            </div>
        </React.Fragment>
        )
    }
}

export default DashboardManagerComponent;