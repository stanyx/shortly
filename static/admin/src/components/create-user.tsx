import * as React from 'react';
import {httpPost, httpGet} from '../utils';

class CreateUserState {
    username:    string;
    email:       string;
    phone:       string;
    password:    string;
    roleID:      number;
    roles:       Array<any>;
}

class CreateUserComponent extends React.Component<any, CreateUserState> {
    constructor(props: any) {
        super(props);
        this.state = {username: '', email: '', phone: '', password: '', roles: [], roleID: 0};
    }
    componentDidMount() {
        httpGet('/api/v1/roles').then((response) => {
            this.setState({roles: response.data.result})
        })
    }
    submit(e: any) {
        e.preventDefault();
        httpPost('/api/v1/users/create', {
            username: this.state.username,
            email:    this.state.email,
            phone:    this.state.phone,
            password: this.state.password,
            roleID:   this.state.roleID,
        })
    }
    render() {
        return (
        <div className="container-fluid">
            <div className="row">
                <div className="col-md-6">
                    <div className="card card-warning">
                        <div className="card-header">
                            <h3 className="card-title">Create User</h3>
                        </div>
                        <div className="card-body">
                            <form role="form">
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Username</label>
                                            <input value={this.state.username}
                                            onChange={(e)=>this.setState({username: e.target.value})}
                                            className="form-control"/>
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Email</label>
                                            <input value={this.state.email}
                                            onChange={(e)=>this.setState({email: e.target.value})}
                                            className="form-control"/>
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Phone</label>
                                            <input value={this.state.phone}
                                            onChange={(e)=>this.setState({phone: e.target.value})}
                                            className="form-control"/>
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Role</label>
                                            <select value={this.state.roleID}
                                            onChange={(e)=>this.setState({roleID: Number(e.target.value)})}
                                            className="form-control">
                                                {this.state.roles.map((r) => {
                                                    return (
                                                        <option value={r.id}>
                                                            {r.name}
                                                        </option>
                                                    )
                                                })}
                                            </select>
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Password</label>
                                            <input value={this.state.password}
                                            onChange={(e)=>this.setState({password: e.target.value})}
                                            className="form-control"/>
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

export default CreateUserComponent;