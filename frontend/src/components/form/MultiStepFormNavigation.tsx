import React from "react";
import { Button } from "@mui/material";
import { FormikValues } from "formik";

interface Props {
    hasPrevious?: boolean;
    isLastStep: boolean;
    next_label?: "Next";
    submit_label?: "Sign Up";
    onBackClick: (values: FormikValues) => void;
}

const MultiStepFormNavigation = ({
    hasPrevious = false, 
    isLastStep, 
    next_label = "Next", 
    submit_label = "Sign Up",
    onBackClick, 
}: Props) => {
    return (
        <div style={{
            display: 'flex',
            marginTop: 25,
            justifyContent: 'space-between',
        }}>
            {hasPrevious ? (
                <Button variant="outlined" type="button" onClick={onBackClick}>
                    Back
                </Button>
            ) : (
                <div /> 
            )}
            <Button variant="outlined" type="submit">
                {isLastStep ? submit_label : next_label}
            </Button>
        </div>
    )
}

export default MultiStepFormNavigation