import * as React from 'react';
import {httpPost} from '../utils';

class CreateLinkState {
    url:        string;
    description: string;
}

class CreateLinkComponent extends React.Component<any, CreateLinkState> {
    constructor(props: any) {
        super(props);
        this.state = {url: '', description: ''};
    }
    submit(e: any) {
        e.preventDefault();
        httpPost('/api/v1/users/links/create', {
            url:         this.state.url,
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
                            <h3 className="card-title">Create Link</h3>
                        </div>
                        <div className="card-body">
                            <form role="form">
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
                                            <label>Long link</label>
                                            <textarea value={this.state.url}
                                            onChange={(e)=>this.setState({url: e.target.value})}
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

export default CreateLinkComponent;