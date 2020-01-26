import * as React from 'react';
import {httpPost, httpGet} from '../utils';

class CreateWebhookState {
    name:    string;
    url:     string;
    events:  Array<string>;
    eventsOptions: Array<string>;
}

class CreateWebhookComponent extends React.Component<any, CreateWebhookState> {
    constructor(props: any) {
        super(props);
        this.state = {name: '', url: '', events: [], eventsOptions: []};
    }
    submit(e: any) {
        e.preventDefault();
        httpPost('/api/v1/webhooks/create', {
            name: this.state.name,
            url:    this.state.url,
            events:    this.state.events,
        })
    }
    render() {
        return (
            <React.Fragment>
            <div className="row">
                <div className="col-md-6">
                    <div className="card card-warning">
                        <div className="card-header">
                            <h3 className="card-title">Create Webhook</h3>
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
                                            <label>URL</label>
                                            <input value={this.state.url}
                                            onChange={(e)=>this.setState({url: e.target.value})}
                                            className="form-control"/>
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Phone</label>
                                            <select value={this.state.events[0]}
                                            onChange={(e)=>this.setState({events: [e.target.value]})}
                                            className="form-control">
                                                <option value="link__created">create link</option>
                                                <option value="link__deleted">delete link</option>
                                                <option value="link__hided">hide link</option>
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

export default CreateWebhookComponent;