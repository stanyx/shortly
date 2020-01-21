import * as React from 'react';
import {httpPost} from '../utils';

class CreateGroupState {
    name:        string;
    description: string;
}

class CreateGroupComponent extends React.Component<any, CreateGroupState> {
    constructor(props: any) {
        super(props);
        this.state = {name: '', description: ''};
    }
    submit(e: any) {
        e.preventDefault();
        httpPost('/api/v1/groups/create', {
            name:        this.state.name,
            description: this.state.description,
        })
    }
    render() {
        return (
        <div className="container-fluid">
            <div className="row">
                <div className="col-md-6">
                    <div className="card card-warning">
                        <div className="card-header">
                            <h3 className="card-title">Create Group</h3>
                        </div>
                        <div className="card-body">
                            <form role="form">
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Name</label>
                                            <input value={this.state.name}
                                            onChange={(e)=>this.setState({name: e.target.value})}
                                            className="form-control"/>
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Description</label>
                                            <textarea value={this.state.description}
                                            onChange={(e)=>this.setState({description: e.target.value})}
                                            className="form-control"></textarea>
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

export default CreateGroupComponent;