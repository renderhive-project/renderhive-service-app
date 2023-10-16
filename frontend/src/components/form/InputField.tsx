import React from "react";
import { TextField } from "@mui/material";
import { useField } from "formik";

interface InputFieldProps {
    name: string;
    label: string;

    fullWidth?: boolean
    disabled?: boolean
    style?: React.CSSProperties;
    value?: string
    type?: string;
}

const InputField = ({ label, ...props }: InputFieldProps ) => {
    const [field, meta] = useField(props);
    
    return (
        <TextField
            fullWidth
            label={label}
            {...field}
            {...props}
            error={meta.touched && Boolean(meta.error)}
            helperText={meta.touched && meta.error}
            style={{ marginTop: 10 }}
        />
    )
}

export default InputField