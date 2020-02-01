import * as React from 'react';
import {httpPost} from '../utils';
import {CardElement, Elements, StripeProvider, injectStripe} from 'react-stripe-elements';


class state {
    planID:         number;
    stripeToken:    string;
    fullName:       string;
    country:        string;
    zipCode:        string;
    isAnnual:       boolean;
}

class PaymentFormComponent extends React.Component<any, state> {

    constructor(props: any) {
        super(props);
        this.state = {
            planID: 0, 
            stripeToken: '', 
            fullName: '', 
            country: '',
            zipCode: '',
            isAnnual: false
        };
    }
    async submit(e: any) {
        e.preventDefault();

        let {token} = await this.props.stripe.createToken({name: "Name"});
        httpPost('/api/v1/billing/apply', {
            planID: Number(this.props.planID),
            fullName: this.state.fullName,
            country: this.state.country,
            zipCode: this.state.zipCode,
            isAnnual: this.state.isAnnual,
            paymentToken: token.id
        }).then(() => {
            console.log("payment success");
        })
    }
    render() {
        return (
        <React.Fragment>
            <div className="row">
                <div className="col-md-6">
                    <div className="card card-warning">
                        <div className="card-header">
                            <h3 className="card-title">Upgrade billing plan</h3>
                        </div>
                        <div className="card-body">
                            <form role="form">
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Full name</label>
                                            <input value={this.state.fullName}
                                            onChange={(e)=>this.setState({fullName: e.target.value})}
                                            className="form-control"/>
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Country</label>
                                            <input value={this.state.country}
                                            onChange={(e)=>this.setState({country: e.target.value})}
                                            className="form-control"/>
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Zip code</label>
                                            <input value={this.state.zipCode}
                                            onChange={(e)=>this.setState({zipCode: e.target.value})}
                                            className="form-control"/>
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Annual payment</label>
                                            <input type="checkbox" checked={this.state.isAnnual}
                                            onChange={(e)=>this.setState({isAnnual: !this.state.isAnnual})}
                                            className="form-control" />
                                        </div>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="col-sm-6">
                                        <CardElement />
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

export default injectStripe(PaymentFormComponent);