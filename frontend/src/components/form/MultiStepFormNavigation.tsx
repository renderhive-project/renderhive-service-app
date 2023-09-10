import React from "react";
import { Button } from "@mui/material";
import { FormikValues } from "formik";

interface Props {
    hasPrevious?: boolean;
    isLastStep: boolean;
    onBackClick: (values: FormikValues) => void;
}

const MultiStepFormNavigation = (props: Props) => {
    return (
        <div style={{
            display: 'flex',
            marginTop: 25,
            justifyContent: 'space-between',
        }}>
            {props.hasPrevious && (
                <Button variant="contained" type="button" onClick={props.onBackClick}>
                    Back
                </Button>
            )}
            <Button variant="contained" type="submit">
                {props.isLastStep ? 'Submit' : 'Next'}
            </Button>
        </div>
    )
}

export default MultiStepFormNavigation