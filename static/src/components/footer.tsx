import * as React from 'react';

const FooterComponent = () => {
    return (
        <footer>
    <div className="footer-top">
        <div className="container">
            <div className="row">
                <div className="col-md-6 col-lg-3 footer_widget">
                    <div className="inner">
                        <h4>About</h4>
                        <p>Urlshortener that works</p>
                    </div>
                </div>
                <div className="col-md-6 col-lg-3 footer_widget">
                    <div className="inner">
                        
                    </div>
                </div>
                <div className="col-md-6 col-lg-3 footer_widget">
                    <div className="inner">
                        
                    </div>
                </div>
                <div className="col-md-6 col-lg-3 footer_widget">
                    <div className="inner">
                        <h4>Address</h4>
                        <h5>Shortly, Inc.</h5>
                        <p><br></br>P: (123) 456-7890</p>
                        <h4>Newsletter</h4>
                        <form action="#" className="nw_form">
                            <input placeholder="Enter your email" type="email"></input>
                            <button><i className="fa fa-paper-plane"></i></button>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    </div>
    <div className="footer-bottom">
        <div className="container">
            <div className="row">
                <div className="col-lg-6">
                    <div className="copyright-txt">
                        © 2020 Shortly. All Rights Reserved.
                    </div>
                </div>
                <div className="col-lg-6 text-right">
                    <div className="footer-nav">
                        <a href="/">Home</a>
                    </div>
                </div>
            </div>
        </div>
    </div>
</footer>
    );
}

export default FooterComponent;