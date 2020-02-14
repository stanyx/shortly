import * as React from 'react';
import {withRouter} from "react-router-dom";
import {httpPost} from '../utils';

class state {
    name:        string;
    description: string;
}

class CreateCampaignComponent extends React.Component<any, state> {
    constructor(props: any) {
        super(props);
        this.state = {name: '', description: ''};
    }
    submit(e: any) {
        e.preventDefault();
        httpPost('/api/v1/campaigns', {
            name:         this.state.name,
            description: this.state.description,
        }).then(() => {
            this.props.history.push("/campaigns");
        })
    }
    render() {
        return (
        <React.Fragment>
            <div className="row">
                <div className="col-md-6">
                    <div className="card card-warning">
                        <div className="card-header">
                            <h3 className="card-title">Create Campaign</h3>
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
        </React.Fragment>
        )
    }
}

export default withRouter(CreateCampaignComponent);