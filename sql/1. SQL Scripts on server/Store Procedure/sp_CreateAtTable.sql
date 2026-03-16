CREATE PROCEDURE [dbo].[sp_CreateAtTable]
(
    @SourceTable VARCHAR(255) 
)
AS
BEGIN
    SET NOCOUNT ON;

    -- Check if the source table name is provided
    IF @SourceTable IS NULL OR @SourceTable = ''
    BEGIN
        RAISERROR('Source table name cannot be null or empty.', 16, 1);
        RETURN;
    END 

    -- Construct the target table name
    DECLARE @TargetTableName NVARCHAR(MAX) =  QUOTENAME('z_'+ @SourceTable+'_at');

    -- Insert data from the source table into the target table
    -- EXEC(N'SELECT * INTO ' + @TargetTableName + ' FROM ' + @SourceTable);
	   
    DECLARE @TableCreateSQL NVARCHAR(MAX) = '';

	 SELECT @TableCreateSQL +=  ' '+QUOTENAME(COLUMN_NAME) + ' varchar(255), '
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_NAME = @SourceTable;
	
    EXEC('CREATE TABLE '+@TargetTableName +' ('+@TableCreateSQL + ')');
	  
	 -- Add the new columns to the target table
    DECLARE @AlterTableSQL NVARCHAR(MAX);
    SET @AlterTableSQL = N'ALTER TABLE ' + @TargetTableName + 
                          ' ADD ' +
                          ' AT_ACTION TEXT, ' +
                          ' IP_ADDRESS TEXT, ' +
						  ' MOTHERBOARD_SERIAL_NO TEXT, ' +
						  ' MACHINE_NAME TEXT, ' +
                          ' AT_DATE TEXT, ' +
                          ' AT_USER_ID TEXT, ' +
                          ' AT_USER TEXT;';

    -- Execute the ALTER TABLE statement to add the new columns
    EXEC(@AlterTableSQL);
END
GO


