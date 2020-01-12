import * as React from 'react';
import axios from 'axios';

interface LoginState {
    password: string;
    email: string;
}

export class LoginComponent extends React.Component<{}, LoginState> {
    constructor(props: any) {
        super(props);
        this.state = {password: "", email: ""};
    }
    submit(e: any) {
        e.preventDefault();
        axios.post('/api/v1/login', {
            password: this.state.password,
            email: this.state.email
        }).then((response: any) => {
            localStorage.setItem('token', response.data.result.token);
            window.location.href = "/";
        });
    }
    render() {
        return (
            <div className="login-block">
                <div className="container">
                    <div className="row">
                        <form className="col-6">
                            <div className="form-title">
                                <h1>Log in</h1>
                                <p>Don't have an account?<a href="app.html?action=register">Sign up</a></p>
                            </div>
                            <fieldset>
                                <label>Email</label>
                                <input value={this.state.email} 
                                    onChange={(e) => this.setState({email: e.target.value})} name="email"/>
                                <label>Password</label>
                                <input value={this.state.password} 
                                    onChange={(e) => this.setState({password: e.target.value})} type="password" name="password"/>
                            </fieldset>
                            <button className="button button-full" onClick={(e) => this.submit(e)}>Login</button>
                            <div className="text-right">
                                <p>
                                    Forgot?
                                </p>
                            </div>
                        </form>
                    </div>
                </div>
            </div> 
        );
    }
}

export default LoginComponent;