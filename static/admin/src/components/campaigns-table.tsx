
import * as React from 'react';
import {Link} from 'react-router-dom';
import {InputSwitch} from 'primereact/inputswitch';

import Paginator from './lib/paginator';
import {httpGet, httpPost, httpDelete} from '../utils';

class state {
    rows: Array<any>;
    total: number;
    limit: number;
    offset: number;
}

class CampaignsTable extends React.Component<any, state> {
    constructor(props: any) {
        super(props);
        this.state = {rows: [], total: 0, limit: 20, offset: 0};
    }
    loadData(limit: number, offset: number) {
        httpGet(`/api/v1/campaigns?limit=${limit}&offset=${offset}`).then((response: any) => {
            this.setState({
                limit: limit,
                offset: offset,
                rows: (response.data.result ? response.data.result: []),
            });
        })
    }
    componentDidMount() {
        this.loadData(this.state.limit, this.state.offset);
    }
    deleteRow(rowID: number) {
        httpDelete(`/api/v1/campaigns/${rowID}`).then(() => {
            this.loadData(this.state.limit, this.state.offset);
        });
    }
    toggleLink(r: any, value: boolean, index: number) {
        if (value) {
            httpPost(`/api/v1/campaigns/${r.id}/start`, {}).then(() => {
                this.state.rows[index].is_active = true;
                this.setState({rows: this.state.rows});
            });
        } else {
            httpPost(`/api/v1/campaigns/${r.id}/stop`, {}).then(() => {
                this.state.rows[index].is_active = false;
                this.setState({rows: this.state.rows});
            });
        }
    }
    render() {
    return (
        <React.Fragment>
            <div className="row">
                <div className="col-12">
                    <Link to="/campaigns/create" className="btn btn-primary">Create</Link>
                    <Link to="/campaigns/dashboard" className="btn btn-success">Dashboard</Link>
                </div>
            </div>
            <div className="row">
                <div className="col-12">
                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Campaigns</h3>
                        </div>
                        <div className="card-body">
                            <div className="dataTables_wrapper dt-bootstrap4">
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
                                                    <th className="sorting">Description</th>
                                                    <th className="sorting">Is Active?</th>
                                                    <th></th>
                                                    <th></th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                            {this.state.rows.map((r: any, index: number) => {
                                                return (
                                                    <tr role="row" className="odd">
                                                        <td className="sorting_1">
                                                            {r.name}
                                                        </td>
                                                        <td>{r.description}</td>
                                                        <td>
                                                            <span>{r.is_active ? 'active': 'not active'}</span>
                                                            <InputSwitch checked={r.is_active}
                                                                onChange={(e: any) => this.toggleLink(r, e.value, index)}>
                                                            </InputSwitch>
                                                        </td>
                                                        <td>
                                                            <Link to={`/campaigns/${r.id}/channels`} className="btn btn-success">
                                                                Channels
                                                            </Link>
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
                                                    <th>Description</th>
                                                    <th>Is Active?</th>
                                                    <th></th>
                                                    <th></th>
                                                </tr>
                                            </tfoot>
                                        </table>
                                    </div>
                                </div>
                                <Paginator 
                                    total={this.state.total} 
                                    offset={this.state.offset} 
                                    limit={this.state.limit} 
                                    loadPage={(page: number) => this.loadData(this.state.limit, (page - 1) * this.state.limit)}/>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </React.Fragment>
        )
    }
}

export default CampaignsTable;