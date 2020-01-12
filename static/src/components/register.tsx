import * as React from 'react';
import axios from 'axios';

interface RegisterState {
    password: string;
    company: string;
    phone: string;
    email: string;
}

export class RegisterComponent extends React.Component<{}, RegisterState> {
    constructor(props: any) {
        super(props);
        this.state = {password: "", company: "", phone: "", email: ""};
    }
    submit(e: any) {
        e.preventDefault();
        axios.post('/api/v1/registration', {
            password: this.state.password,
            company: this.state.company,
            phone: this.state.phone,
            email: this.state.email
        }).then((response: any) => {
            window.location.href = "app.html?action=login";
        });
    }
    render() {
        return (
            <div className="register-block">
                <div className="container">
                    <div className="row">
                        <form className="col-6">
                            <div className="form-title">
                                <h1>Sign up</h1>
                                <p>Already have an account?<a href="app.html?action=login">Log In</a></p>
                            </div>
                            <fieldset>
                                <label>Company</label>
                                <input value={this.state.company} 
                                    onChange={(e) => this.setState({company: e.target.value})} name="company"/>
                                <label>Phone</label>
                                <input value={this.state.phone} 
                                    onChange={(e) => this.setState({phone: e.target.value})} name="phone"/>
                                <label>Email</label>
                                <input value={this.state.email} 
                                    onChange={(e) => this.setState({email: e.target.value})} name="email"/>
                                <label>Password</label>
                                <input value={this.state.password} 
                                    onChange={(e) => this.setState({password: e.target.value})} type="password" name="password"/>
                            </fieldset>
                            <button className="button button-full" onClick={(e) => this.submit(e)}>Register</button>
                            <div>
                                <p>
                                    By creating an account, you agree to <br/>
                                    Shortly's Terms Of Service and Privacy Policy
                                </p>
                            </div>
                        </form>
                    </div>
                </div>
            </div> 
        );
    }
}

export default RegisterComponent;