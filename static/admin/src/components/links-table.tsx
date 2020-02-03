
import * as React from 'react';
import {Link} from 'react-router-dom';
import {InputSwitch} from 'primereact/inputswitch';

import {httpGet, httpPost, httpDelete} from '../utils';

class LinkTableState {
    rows: Array<any>;
}

class LinksTable extends React.Component<any, LinkTableState> {
    constructor(props: any) {
        super(props);
        this.state = {rows: []};
    }
    loadData() {
        httpGet('/api/v1/users/links').then((response: any) => {
            this.setState({rows: (response.data.result ? response.data.result: [])});
        })
    }
    componentDidMount() {
        this.loadData();
    }
    deleteRow(linkID: number) {
        httpDelete(`/api/v1/users/links/${linkID}`).then(() => {
            this.loadData();
        });
    }
    toggleLink(r: any, value: boolean, index: number) {
        if (value) {
            httpPost(`/api/v1/links/${r.id}/activate`, {}).then(() => {
                this.state.rows[index].is_active = true;
                this.setState({rows: this.state.rows});
            });
        } else {
            httpPost(`/api/v1/links/${r.id}/hide`, {}).then(() => {
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
                    <Link to="/links/create" className="btn btn-primary">Create</Link>
                </div>
            </div>
            <div className="row">
                <div className="col-12">
                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Links</h3>
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
                                                    <th className="sorting_asc">Short link</th>
                                                    <th className="sorting">Long link</th>
                                                    <th className="sorting">Is Active</th>
                                                    <th></th>
                                                    <th></th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                            {this.state.rows.map((r: any, index: number) => {
                                                return (
                                                    <tr role="row" className="odd">
                                                        <td className="sorting_1">
                                                            <a href={"http://" + location.host + "/" + r.short}>{r.short}</a>
                                                        </td>
                                                        <td>{r.long}</td>
                                                        <td>
                                                            <span>{r.is_active ? 'active': 'not active'}</span>
                                                            <InputSwitch checked={r.is_active}
                                                                onChange={(e: any) => this.toggleLink(r, e.value, index)}>
                                                            </InputSwitch>
                                                        </td>
                                                        <td>
                                                            <Link to={`/links/${r.id}/tags`} className="btn btn-success">
                                                                Tags
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
                                                    <th>Short link</th>
                                                    <th>Long link</th>
                                                    <th>Is Active?</th>
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

export default LinksTable;