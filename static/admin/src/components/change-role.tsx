import * as React from 'react';
import {httpPut, httpGet} from '../utils';

class ChangeRoleState {
    roleID: number;
    roles: Array<any>;
}

class ChangeRoleComponent extends React.Component<any, ChangeRoleState> {
    constructor(props: any) {
        super(props);
        this.state = {roles: [], roleID: 0};
    }
    componentDidMount() {
        httpGet("/api/v1/roles").then((response) => {
            this.setState({roles: response.data.result || []});
        });
    }
    submit(e: any) {
        e.preventDefault();
        httpPut("/api/v1/roles/change", {userID: Number(this.props.userID), roleID: this.state.roleID});
    }
    render() {
        return (
        <div className="container-fluid">
            <div className="row">
                <div className="col-md-6">
                    <div className="card card-warning">
                        <div className="card-header">
                            <h3 className="card-title">Change user role</h3>
                        </div>
                        <div className="card-body">
                            <form role="form">
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <select value={this.state.roleID}
                                                onChange={(e)=>this.setState({roleID: Number(e.target.value)})}>
                                                {this.state.roles.map((r) => {
                                                    return (
                                                        <option value={r.id}>{r.name}</option>
                                                    );
                                                })}
                                            </select>
                                        </div>
                                    </div>
                                </div>
                            </form>
                            <button className="btn btn-danger" onClick={(e) => this.submit(e)}>Submit</button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        )
    }
}

export default ChangeRoleComponent;