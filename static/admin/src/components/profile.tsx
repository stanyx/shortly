import * as React from 'react';
import {httpGet} from '../utils';

class ProfileComponentState {
    username: string;
    company: string;
    roleName: string;
    billingPlan: string;
    billingPlanFee: string;
    billingPlanExpiredAt: string;
    billingUsage: Array<any>;
    plansAvailable: Array<any>;
    upgradedPlan: string;
}


class ProfileComponent extends React.Component<any, ProfileComponentState> {
    planSelect: React.RefObject<any>;
    constructor(props: any) {
        super(props)
        this.planSelect = React.createRef();
        this.state = {
            username: "-", 
            company: "-",
            roleName: "-",
            billingPlan: "-", 
            billingPlanFee: "-",
            billingPlanExpiredAt: "-",
            upgradedPlan: "",
            billingUsage: [],
            plansAvailable: []
        };
    }

    componentDidMount() {

        httpGet("/api/v1/profile").then((response) => {
            const d = response.data.result;
            this.setState({
                username: d.username,
                company: d.company,
                roleName: d.roleName,
                billingPlan: d.billingPlan,
                billingPlanFee: d.billingPlanFee,
                billingPlanExpiredAt: d.billingPlanExpiredAt,
                billingUsage: d.billingUsage,
                plansAvailable: d.plansAvailable,
            });

        })
    }

    upgradePlan(e: any) {
        e.preventDefault();
    }

    render() {
        return (
            <React.Fragment>
                <div className="row">
                    <div className="col-6">
                        <div className="card">
                            <div className="card-header">
                                <h3 className="card-title">User information</h3>
                            </div>
                            <div className="card-body">
                                <dl>
                                    <dt>Company</dt>
                                    <dd>{this.state.company}</dd>
                                    <dt>Username</dt>
                                    <dd>{this.state.username}</dd>
                                    <dt>Role</dt> 
                                    <dd>{this.state.roleName}</dd>    
                                </dl>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="row">
                    <div className="col-6">
                        <div className="card">
                            <div className="card-header">
                                <h3 className="card-title">Billing information</h3>
                            </div>
                            <div className="card-body">
                                <dl>
                                    <dt>Billing Plan</dt>
                                    <dd>{this.state.billingPlan}</dd>
                                    <dt>Expired at</dt>
                                    <dd>{this.state.billingPlanExpiredAt}</dd>
                                    <dt>Upgrade</dt> 
                                    <dd>
                                        <select ref={this.planSelect} name="billingUpgrade">
                                            {this.state.plansAvailable.map((p) => {
                                                return (
                                                    <option value={p.id}>{p.name}</option>
                                                )})
                                            }
                                        </select>
                                        <button className="btn btn-warning" onClick={(e) => this.upgradePlan(e)}>
                                            Upgrade
                                        </button>
                                    </dd>    
                                </dl>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="row">
                    <div className="col-6">
                        <div className="card">
                            <div className="card-header">
                                <h3 className="card-title">Billing usage</h3>
                            </div>
                            <div className="card-body">
                                <dl>
                                    {this.state.billingUsage.map((counter) => {
                                        return (
                                            <React.Fragment>
                                                <dt>
                                                    {counter.name}
                                                    <div>
                                                        <label className="btn btn-primary">Current: <span className="badge badge-light">{counter.value}</span></label>
                                                        <label className="btn btn-warning">Limit: <span className="badge badge-light">{counter.limit}</span></label>
                                                    </div>
                                                </dt>
                                                <dd>
                                                    <div className="progress progress-xxs">
                                                        <div className="progress-bar progress-bar-danger progress-bar-striped" role="progressbar" aria-valuenow={(counter.value / counter.limit) * 100.0} aria-valuemin={0.0} aria-valuemax={100.0}>
                                                        </div>
                                                    </div>
                                                </dd>
                                            </React.Fragment>
                                        )
                                    })}
                                </dl>
                            </div>
                        </div>
                    </div>
                </div>
            </React.Fragment>
        )
    }
}

export default ProfileComponent