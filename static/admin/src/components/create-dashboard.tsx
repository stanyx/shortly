import * as React from 'react';
import {httpPost} from '../utils';

class state {
    name:        string;
    description: string;
    gridType:    string;
}

class CreateDashboardComponent extends React.Component<any, state> {
    constructor(props: any) {
        super(props);
        this.state = {name: '', description: '', gridType: ''};
    }
    submit(e: any) {
        e.preventDefault();

        const dim = this.state.gridType.split(",");

        const width = Number(dim[0]);
        const height = Number(dim[1]);

        httpPost('/api/v1/dashboards/create', {
            name:        this.state.name,
            description: this.state.description,
            width:       width,
            height:      height
        })
    }
    render() {
        return (
        <React.Fragment>
            <div className="row">
                <div className="col-md-6">
                    <div className="card card-warning">
                        <div className="card-header">
                            <h3 className="card-title">Create Dashboard</h3>
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
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Grid Type</label>
                                            <select value={this.state.gridType}
                                            onChange={(e)=>this.setState({gridType: e.target.value})}
                                            className="form-control">
                                                <option value="2,2">2 X 2</option>
                                                <option value="2,3">2 X 3</option>
                                                <option value="2,4">2 X 4</option>
                                                <option value="3,2">3 X 2</option>
                                                <option value="3,3">3 X 3</option>
                                                <option value="3,4">3 X 4</option>
                                                <option value="4,2">4 X 2</option>
                                                <option value="4,3">4 X 3</option>
                                                <option value="4,4">4 X 4</option>
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
        </React.Fragment>
        )
    }
}

export default CreateDashboardComponent;