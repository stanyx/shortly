
import * as React from 'react';
import {Link} from 'react-router-dom';
import {httpGet, httpDelete, httpPost} from '../utils';

class WebhookTableState {
    rows: Array<any>;
}

class WebhookTableComponent extends React.Component<any, WebhookTableState> {
    constructor(props: any) {
        super(props);
        this.state = {rows: []};
    }
    loadData() {
        httpGet('/api/v1/webhooks').then((response: any) => {
            this.setState({rows: (response.data.result ? response.data.result: [])});
        })
    }
    componentDidMount() {
        this.loadData();
    }
    deleteRow(rowID: number) {
        httpDelete(`/api/v1/webhooks/${rowID}`).then(() => {
            this.loadData();
        })
    }
    toggleWebhook(row: any) {
        if (!row.active) {
            httpPost(`/api/v1/webhooks/${row.id}/enable`, {}).then(() => {
                this.loadData();
            })
        } else {
            httpPost(`/api/v1/webhooks/${row.id}/disable`, {}).then(() => {
                this.loadData();
            })
        }
    }
    render() {
    return (
        <React.Fragment>
            <div className="row">
                <div className="col-12">
                    <Link to="/webhooks/create" className="btn btn-primary">Create</Link>
                </div>
            </div>
            <div className="row">
                <div className="col-12">
                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Webhooks</h3>
                        </div>
                        <div className="card-body">
                            <div id="example2_wrapper" className="dataTables_wrapper dt-bootstrap4">
                                <div className="row">
                                    <div className="col-sm-12 col-md-6"></div>
                                    <div className="col-sm-12 col-md-6"></div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-12">
                                        <table id="example2" className="table table-bordered table-hover dataTable" role="grid" aria-describedby="example2_info">
                                            <thead>
                                                <tr role="row">
                                                    <th className="sorting_asc">Name</th>
                                                    <th className="sorting">URL</th>
                                                    <th className="sorting">Events</th>
                                                    <th className="sorting">Is active?</th>
                                                    <th></th>
                                                    <th></th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                            {this.state.rows.map((r: any) => {
                                                return (
                                                    <tr role="row" className="odd">
                                                        <td className="sorting_1">
                                                            {r.name}
                                                        </td>
                                                        <td>{r.url}</td>
                                                        <td>{r.events}</td>
                                                        <td>{r.active ? 'active': 'not active'}</td>
                                                        <td>
                                                            <button onClick={(e) => this.toggleWebhook(r)} className="btn btn-primary">
                                                                {r.active ? 'Deactivate': 'Activate'}
                                                            </button>
                                                        </td>
                                                        <td>
                                                            <button onClick={(e) => this.deleteRow(r.id)} className="btn btn-danger">
                                                                Delete
                                                            </button>
                                                        </td>
                                                    </tr>
                                                )
                                            })}
                                            </tbody>
                                            <tfoot>
                                                <tr>
                                                    <th>Name</th>
                                                    <th>URL</th>
                                                    <th>Events</th>
                                                    <th>Is active?</th>
                                                    <th></th>
                                                    <th></th>
                                                </tr>
                                            </tfoot>
                                        </table>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-12 col-md-5">
                                        <div className="dataTables_info" id="example2_info" role="status" aria-live="polite">
                                            Showing 1 to 10 of - entries
                                        </div>
                                    </div>
                                    <div className="col-sm-12 col-md-7">
                                        <div className="dataTables_paginate paging_simple_numbers" id="example2_paginate">
                                            <ul className="pagination">
                                                <li className="paginate_button page-item previous disabled" id="example2_previous">
                                                    <a href="#" className="page-link">Previous</a>
                                                </li>
                                                <li className="paginate_button page-item active">
                                                    <a href="#" className="page-link">1</a>
                                                </li>
                                                <li className="paginate_button page-item ">
                                                    <a href="#" className="page-link">2</a>
                                                </li>
                                                <li className="paginate_button page-item ">
                                                    <a href="#" className="page-link">3</a>
                                                </li>
                                                <li className="paginate_button page-item ">
                                                    <a href="#" className="page-link">4</a>
                                                </li>
                                                <li className="paginate_button page-item ">
                                                    <a href="#" className="page-link">5</a>
                                                </li>
                                                <li className="paginate_button page-item ">
                                                    <a href="#" className="page-link">6</a>
                                                </li>
                                                <li className="paginate_button page-item next" id="example2_next">
                                                    <a href="#" className="page-link">Next</a>
                                                </li>
                                            </ul>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </React.Fragment>
        )
    }
}

export default WebhookTableComponent;