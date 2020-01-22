
import * as React from 'react';
import {httpGet, httpPost} from '../utils';

class PermissionTableState {
    roleID: string;
    rows: Array<any>;
}

class PermissionTable extends React.Component<any, PermissionTableState> {
    constructor(props: any) {
        super(props);
        this.state = {rows: [], roleID: this.props.roleID || ""};
    }
    componentDidMount() {
        httpGet(`/api/v1/permissions/${this.state.roleID}`).then((response: any) => {
            this.setState({rows: (response.data.result ? response.data.result: [])});
        })
    }
    grantAccess(r: any) {
        httpPost(`/api/v1/roles/grant`, {roleID: Number(this.state.roleID), resource: r.url, method: r.method});
    }
    revokeAccess(r: any) {
        httpPost(`/api/v1/roles/revoke`, {roleID: Number(this.state.roleID), resource: r.url, method: r.method});
    }
    render() {
    return (
        <section className="content">
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
                                                    <th className="sorting_asc">Resource</th>
                                                    <th className="sorting">Method</th>
                                                    <th className="sorting"></th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                            {this.state.rows.map((r: any) => {
                                                return (
                                                    <tr role="row" className="odd">
                                                        <td className="sorting_1">
                                                            {r.url}
                                                        </td>
                                                        <td>{r.method}</td>
                                                        <td>
                                                        {r.accessGranted ? (
                                                            <button onClick={(e) => this.revokeAccess(r)} className="btn btn-success">
                                                                Revoke
                                                            </button>
                                                        ): (
                                                            <button onClick={(e) => this.grantAccess(r)} className="btn btn-danger">
                                                                Grant
                                                            </button>
                                                        )}
                                                        </td>
                                                    </tr>
                                                )
                                            })}
                                            </tbody>
                                            <tfoot>
                                                <tr>
                                                    <th>Resource</th>
                                                    <th>Method</th>
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
        </section>
        )
    }
}

export default PermissionTable;