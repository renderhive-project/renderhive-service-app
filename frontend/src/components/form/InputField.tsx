import React from "react";
import { BaseTextFieldProps, TextField, TextFieldProps, TextFieldVariants } from "@mui/material";
import { FieldConfig, FieldHookConfig, useField } from "formik";

interface InputFieldProps {
    name: string;
    label: string;

    fullWidth?: boolean
    disabled?: boolean
    style?: React.CSSProperties;
    value?: string
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