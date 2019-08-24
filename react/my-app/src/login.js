import React from 'react';


function ErrorDisplay() {
    return (<div class="w3-panel w3-red">
        <h3>Unable to login!</h3>
        <p>Username/password is incorrect</p>
    </div>);
}

function OtherError() {
    return (
        <div class="w3-panel w3-yellow">
            <h3>Oh no!</h3>
            <p>I wasn't able to connect to the API.</p>
        </div>
    );
}

export function CheckAuth() {
    fetch('http://localhost:8000/s/refresh', {
        method: 'GET',
        credentials: 'include',

    })
        .then((res) => {
            if (res.status === 401) {
                console.log("Returning false")
                return false
            } else {
                console.log("Returning true")
                return true
            }

        },
            (error) => {

                return false

            }
        )

        console.log("Reached end of function")
}

export class LoginForm extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            username: '',
            password: '',
            error: false,
            connectionError: false
        };
        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }



    handleChange(event) {
        this.setState({
            [event.target.name]: event.target.value
        });
    }

    handleSubmit(event) {

        fetch('http://localhost:8000/login', {
            method: 'POST',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({
                username: this.state.username,
                password: this.state.password,
            })
        })
            .then((res) => {
                if (res.status === 401) {
                    this.setState({ error: true })
                    localStorage.setItem('login', "false");
                } else if (res.status === 200) {
                    console.log("logged in");
                    localStorage.setItem('login', "true");
                    window.location.reload();
                }

            },
                (error) => {
                    this.setState({
                        connectionError: true
                    });
                }
            )

        event.preventDefault();
    }

    render() {

        return (
            <div class="w3-display-topmiddle">
                {this.state.error &&
                    <ErrorDisplay />
                }
                {this.state.connectionError &&
                    <OtherError />
                }
                <form class="w3-container" onSubmit={this.handleSubmit}>
                    <div class="w3-section">
                        <label><b>Username</b></label>
                        <input class="w3-input w3-border w3-margin-bottom" name="username" type="text" value={this.state.username} onChange={this.handleChange} placeholder="Username" />

                        <label><b>Password</b></label>
                        <input class="w3-input w3-border w3-margin-bottom" name="password" type="password" value={this.state.password} onChange={this.handleChange} placeholder="Password" />
                        <button class="w3-button w3-block w3-green w3-section w3-padding" type="submit">Login</button>
                    </div>
                </form>
            </div>

        );
    }
}


export default LoginForm