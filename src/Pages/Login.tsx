import './Styles/login.css'

import { NavLink } from "react-router-dom";

import { useState } from "react";
import { useForm } from '@tanstack/react-form'
import motchi_pixel_logo from "../assets/motchi_pixel_logo.svg"
import { MdOutlineVisibility } from "react-icons/md";
import { MdOutlineVisibilityOff } from "react-icons/md";
import { useMutation } from '@tanstack/react-query';
import axios from 'axios';


interface RegistrationFormValues {
    username: string;
    password: string;
}

export function Login() {

    const [showPassword, setShowPassword] = useState(false);

    const togglePasswordVisibility = () => {
        setShowPassword(!showPassword);
    };

    const CLIENT_ID = import.meta.env.VITE_OAUTH2_CLIENT_ID ?? ""
    const CLIENT_SECRET = import.meta.env.VITE_OAUTH2_CLIENT_SECRET ?? ""

    const mutation = useMutation({
        mutationFn: (newUser: RegistrationFormValues) => {
            const params = new URLSearchParams();
            params.append('grant_type', 'password');
            params.append('username', newUser.username);
            params.append('password', newUser.password);
            // include client credentials if available (set via env for dev)
            if (CLIENT_ID) params.append('client_id', CLIENT_ID);
            if (CLIENT_SECRET) params.append('client_secret', CLIENT_SECRET);

            console.log(params.toString());
            return axios.post('/api/token', params, {
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded'
                }
            });
        },
        onSuccess: (data) => {
            // store full token JSON (access_token, expires_in, extensions like user_id/pet_id)
            console.log(data);
            localStorage.setItem('auth_token', data.data.access_token);
            window.location.href = '/';
            console.log('Login successful, token saved.');
        },
        onError: (err) => {
            alert(err);
            console.error('Login error', err);
        }
    });

    const form = useForm({
        defaultValues: {
            username: '',
            password: '',
        } as RegistrationFormValues,
        onSubmit: ({ value }) => {
            const user: RegistrationFormValues = {
                username: value.username,
                password: value.password,
            }
            mutation.mutate(user);
        },
    })

    return (
        <>
            <main className='login-cont'>
                <header>
                    <img src={`${motchi_pixel_logo}`} alt="logo for motchi" className="logo" />
                    <h1>Sign In</h1>
                </header>
                <form className='formContainer'
                    onSubmit={(e) => {e.preventDefault(); form.handleSubmit()}}
                >
                    <div className='form'>
                        {/* Username */}
                        <form.Field
                            name='username'
                            validators={{
                                onChange: ({ value }) => {
                                    return value.trim() === "" ? "Enter username" : undefined
                                }
                            }}
                            children={(field) => (
                                <div className='field'>
                                    <input
                                        placeholder='Username'
                                        type='text'
                                        id={field.name}
                                        value={field.state.value}
                                        onBlur={field.handleBlur}
                                        onChange={(e) => field.handleChange(e.target.value)}
                                    />
                                    {field.state.meta.errors.length > 0 && (
                                        <p className='warning'>{field.state.meta.errors.join(", ")}</p>
                                    )}
                                </div>
                            )}
                        />
                        {/* Password */}
                        <form.Field
                            name='password'
                            validators={{
                                onChange: ({ value }) => {
                                    return value.trim() === "" ? "Enter password" : undefined
                                }
                            }}
                            children={(field) => (
                                <>
                                    <div className='field password-cont'>
                                        <input type={showPassword ? 'text' : 'password'}
                                            placeholder='Password'
                                            id={field.name}
                                            name={field.name}
                                            value={field.state.value}
                                            onBlur={field.handleBlur}
                                            onChange={(e) => field.handleChange(e.target.value)}
                                        />
                                        <button className='icon-button'
                                            type="button"
                                            onClick={togglePasswordVisibility}>
                                            {showPassword ? <MdOutlineVisibilityOff /> : <MdOutlineVisibility />}
                                        </button>
                                    </div>
                                    <div>
                                        {field.state.meta.errors.length > 0 && (
                                            <p className='warning'>{field.state.meta.errors.join(", ")}</p>
                                        )}
                                    </div>
                                </>
                            )}
                        />
                    </div>
                    <button type='submit' className='CTA'>Log in</button>

                    {mutation.isError && (
                        <p className="warning">Login failed. Check credentials.</p>
                    )}

                    <div className="register">
                        <p>Don't have an account?</p>
                        <NavLink to='/sign-up' className='register-link'>
                            <p>Register</p>
                        </NavLink>
                    </div>
                </form>
            </main>
        </>
    );
}