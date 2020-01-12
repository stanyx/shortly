import * as React from 'react';

const HeaderComponent = () => {
    return (
        <header>
            <div className="container">
                <div className="row">
                    <div className="col-6 col-sm-3 logo-column">
                        <a href="/" className="logo">
                            <img src="assets/img/logo.png" alt="logo"></img>
                        </a>
                    </div>
                    <div className="col-6 col-sm-9 nav-column clearfix">
                        <div className="right-nav">
                            <span className="search-icon fa fa-search"></span>
                            <form action="#" className="search-form">
                                <input type="search" placeholder="search now"></input>
                                <button type="submit"><i className="fa fa-search"></i></button>
                            </form>
                            <div className="header-social">
                                <a href="/" className="fa fa-facebook"></a>
                                <a href="/" className="fa fa-twitter"></a>
                                <a href="/" className="fa fa-github"></a>
                            </div>
                        </div>
                        <nav id="menu" className="d-none d-lg-block">
                            <ul>
                                <li><a href="/">Get started</a></li>
                                <li><a href="/">What is Shortly</a></li>
                                <li><a href="/">FAQ</a></li>
                                <li><a href="app.html?action=login" className="button">Sign in</a></li>
                                <li><a href="app.html?action=register" className="button-2">Sign up</a></li>
                            </ul>
                        </nav>
                    </div>
                </div>
            </div>
        </header>
    );
}

export default HeaderComponent;