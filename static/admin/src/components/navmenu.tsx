import * as React from 'react';
import {Link} from 'react-router-dom';

const NavMenu = () => {
    return (
        <nav className="mt-2">
        <ul className="nav nav-pills nav-sidebar flex-column" data-widget="treeview" role="menu" data-accordion="false">
        <li className="nav-item">
            <Link to="/admin" className="nav-link active">
            <i className="nav-icon fas fa-tachometer-alt"></i>
            <p>
                Dashboard
                <i className="right fas fa-angle-left"></i>
            </p>
            </Link>
        </li>
        <li className="nav-item">
            <Link to="/links" className="nav-link">
            <i className="nav-icon fas fa-th"></i>
            <p>
                Links
            </p>
            </Link>
        </li>
        <li className="nav-item">
            <Link to="/roles" className="nav-link">
            <i className="nav-icon fas fa-th"></i>
            <p>
                Roles
            </p>
            </Link>
        </li>
        <li className="nav-item">
            <Link to="/users" className="nav-link">
            <i className="nav-icon fas fa-th"></i>
            <p>
                Users
            </p>
            </Link>
        </li>
        <li className="nav-item">
            <Link to="/groups" className="nav-link">
            <i className="nav-icon fas fa-th"></i>
            <p>
                Groups
            </p>
            </Link>
        </li>
        <li className="nav-item">
            <Link to="/webhooks" className="nav-link">
            <i className="nav-icon fas fa-th"></i>
            <p>
                Webhooks
            </p>
            </Link>
        </li>
        </ul>
    </nav>
    )
}

export default NavMenu;