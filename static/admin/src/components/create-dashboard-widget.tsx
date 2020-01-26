import * as React from 'react';
import {httpPost} from '../utils';

class state {
    title: string;
    type: string;
    posX: number;
    posY: number;
    span: number;
    dataURL: string;
}

class CreateDashboardWidgetComponent extends React.Component<any, state> {
    constructor(props: any) {
        super(props);
        this.state = {
            title: '', 
            type: '', 
            posX: this.props.posX, 
            posY: this.props.posY, 
            dataURL: '',
            span: 1,
        };
    }
    submit(e: any) {
        e.preventDefault();

        if (this.state.span > this.props.maxSpan) {
            console.log("span must be equal or lower than maxSpan");
            return;
        }

        httpPost(`/api/v1/dashboards/${this.props.dashboardID}/widgets`, {
            title:   this.state.title,
            type:    this.state.type,
            posX:    Number(this.state.posX),
            posY:    Number(this.state.posY),
            span:    Number(this.state.span),
            dataURL: this.state.dataURL,
        })
    }
    render() {
        return (
        <React.Fragment>
            <div className="row">
                <div className="col-md-6">
                    <div className="card card-warning">
                        <div className="card-header">
                            <h3 className="card-title">Add Widget</h3>
                        </div>
                        <div className="card-body">
                            <form role="form">
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Title</label>
                                            <input value={this.state.title}
                                            onChange={(e)=>this.setState({title: e.target.value})}
                                            className="form-control"/>
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>DataURL</label>
                                            <input value={this.state.dataURL}
                                            onChange={(e)=>this.setState({dataURL: e.target.value})}
                                            className="form-control"/>
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Widget Type</label>
                                            <select value={this.state.type}
                                            onChange={(e)=>this.setState({type: e.target.value})}
                                            className="form-control">
                                                <option value="line_chart">Line chart</option>
                                                <option value="counter">Counter</option>
                                                <option value="text_block">Text block</option>
                                            </select>
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Span</label>
                                            <input type="number" value={this.state.span}
                                            onChange={(e)=>this.setState({span: Number(e.target.value)})}
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
        </React.Fragment>
        )
    }
}

export default CreateDashboardWidgetComponent;