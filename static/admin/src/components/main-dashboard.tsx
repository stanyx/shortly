import * as React from 'react';
import {httpPost, httpGet} from '../utils';
import {Chart} from 'primereact/chart';

class widgetState {
    value: any;
}

class Widget extends React.Component<any, widgetState> {
    constructor(props: any) {
        super(props);
        this.state = {value: null};
    }
    componentDidMount() {
        if (this.props.dataURL) {
            httpGet(this.props.dataURL).then((response) => {
                this.setState({value: response.data.result});
            });
        }
    }
    render() {
        if (this.props.type == "counter") {
            return (
                <div className="small-box bg-info">
                    <div className="inner">
                        <h3>{this.state.value || '-'}</h3>
                        <p>{this.props.title}</p>
                    </div>
                    <a href="#" className="small-box-footer">
                    
                    </a>
                </div>
            )
        } else if (this.props.type == "text_block") {
            return (
                <span>{this.state.value}</span>
            )
        } else if (this.props.type == "line_chart") {
            return (
                <div className="card">
                    <div className="card-header">
                        <h3>{this.props.title}</h3>
                    </div>
                    <div className="card-body">
                        <div className="dashboard-chart-container">
                            <Chart type="bar" height="400" options={{maintainAspectRatio: false}} data={this.state.value || []} />
                        </div>
                    </div>
                </div>
            )
        } else {
            return null
        }
    }
}

class state {
    dashboards: Array<any>;
    widgets: Array<any>;
    colWidth: number;
    activeDashboard: any;
    links: Array<any>;
    linkStatistics: any;
}

class MainDashboardComponent extends React.Component<any, state> {
    constructor(props: any) {
        super(props);
        this.state = {
            dashboards: [], 
            widgets: [], 
            colWidth: 0, 
            activeDashboard: null,
            links: [],
            linkStatistics: {referers: {}, locations: {}, clicks: {}}
        };
    }
    componentDidMount() {
        httpGet(`/api/v1/dashboards`).then((response) => {
            this.setState({dashboards: (response.data.result || [])});

            if (this.state.dashboards.length > 0) {
                this.showData(this.state.dashboards[0]);
            }
        })
        httpGet(`/api/v1/users/links`).then((response) => {
            this.setState({links: (response.data.result || [])});
        })
    }
    showData(dashboard: any) {
        this.setState({activeDashboard: dashboard});
        httpGet(`/api/v1/dashboards/${dashboard.id}/widgets`).then((response) => {

            var grid: Array<any> = [];

            for (let i = 0; i < dashboard.height; i++) {
                const column = [];
                for (let h = 0; h < dashboard.width; h++) {
                    column.push({});
                }
                grid.push(column);
            }

            const widgets = (response.data.result || []);

            for (let w of widgets) {
                grid[w.posY][w.posX] = w;
                grid[w.posY].splice(w.posX + 1, w.span - 1);
            }

            this.setState({widgets: grid, colWidth: Number(12/dashboard.width)});
        })
    }
    showLinkStatistics(l: any) {
        httpGet(`/api/v1/users/links/${l.short}/stat`).then((response) => {
            this.setState({linkStatistics: response.data.result});
        })
    }
    render() {
        return (
        <div className="container-fluid">
            <div className="row">
                <div className="col-12">
                    <ul className="nav nav-pills">
                        {this.state.dashboards.map((r) => {
                            return (
                                <li className="nav-item">
                                    <a href="" onClick={(e) => this.showData(r)} 
                                    className={"text-sm-center nav-link " + (this.state.activeDashboard === r ? 'active': '')}>
                                        {r.name}
                                    </a>
                                </li>
                            )
                        })}
                    </ul>
                </div>
            </div>

            {this.state.widgets.map((row) => {
                return (
                <div className="row">
                    {
                        row.map((w: any) => {
                            return (
                                <div className={"col-" + (this.state.colWidth * (w.span || 1))}>
                                    <Widget type={w.type} title={w.title} dataURL={w.dataUrl}/>
                                </div>
                            )
                        })
                    }     
                </div>
                )
            })}

            <div className="row">
                <div className="col-6">
                    <div className="card card-outline card-primary">
                        <div className="card-header">
                            <h3>Links</h3>
                        </div>
                        <div className="card-body">
                            <ul className="list-group">
                                {this.state.links.map((l: any) => {
                                    return (
                                        <li onClick={(e) => this.showLinkStatistics(l)} className="list-group-item">
                                            <div className="long-link"><strong>{l.long}</strong></div>
                                            <span className="short-link">{l.short}</span>
                                            <i className="badge">
                                                {l.clicks}
                                            </i>
                                        </li>
                                    )
                                })}
                            </ul>
                        </div>
                    </div>
                </div>
                <div className="col-6">
                    <div className="card card-outline card-primary">
                        <div className="card-header">
                            <h3>Statistics</h3>
                        </div>
                        <div className="card-body">
                            <div className="container">
                                <div className="row">
                                    <div className="col-12">
                                        <Chart type="bar" height="300" options={{title: {
                                            display: true,
                                            text: 'Clicks',
                                            fontSize: 16
                                        }, maintainAspectRatio: false}} data={this.state.linkStatistics.clicks}>

                                        </Chart>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-6">
                                        <Chart type="pie" options={{title: {
                                            display: true,
                                            text: 'Referrers',
                                            fontSize: 16
                                        }}} data={this.state.linkStatistics.referrers}></Chart>
                                    </div>
                                    <div className="col-6">
                                        <Chart type="pie" options={{title: {
                                            display: true,
                                            text: 'Locations',
                                            fontSize: 16
                                        }}} data={this.state.linkStatistics.locations}></Chart>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

        </div>
        )
    }
}

export default MainDashboardComponent;