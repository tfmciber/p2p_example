import netCDF4 as nc
import pandas as pd
import matplotlib.pyplot as plt

# Open the netCDF4 file
files = ["data.nc","data2.nc"]
fig, ax = plt.subplots(1,len(files))
i=0
for file in files:

    nc_file =nc.Dataset(file)

    # Get the variables and index
    time_var = nc_file.variables['time']
    t2m_var = nc_file.variables['t2m']


    # Convert the time values to a standard format
    time_units = time_var.units

    time_calendar = time_var.calendar
    time_data = (nc.num2date(time_var[:], units=time_units, calendar=time_calendar))
    for j in range(len(time_data)):
        
        datestr =str(time_data[j])
        time_data[j] = pd.to_datetime(datestr, format='%Y-%m-%d %H:%M:%S')

    # Extract the data and create a pandas dataframe
    print(time_var)
    df = pd.DataFrame({'t2m': t2m_var[:,0,0,0]-273.15},
                    index=(time_data))

    #remove last year of data


    # Plot the variables

    # get the means for each year
    df_mean = df.groupby(df.index.year).mean()
    #remove last year of data
    df_mean = df_mean[:-1]
    df_max = df.groupby(df.index.year).max()
    df_max = df_max[:-1]
    df_min = df.groupby(df.index.year).min()
    df_min = df_min[:-1]
    nc_file.close()

    #plot max mean and min

    ax[i].plot(df_mean.index, df_mean['t2m'], color='blue', label='mean')
    ax[i].plot(df_max.index, df_max['t2m'], color='red', label='max')
    ax[i].plot(df_min.index, df_min['t2m'], color='green', label='min')
    ax[i].xlim(-20,40)
    i=i+1

plt.legend()
# Set the plot title and axes labels
plt.title('Valores de Temperatura a 2 metros del suelo en N[41 40] E[-3.3 -3.2] a las 16H')
plt.xlabel('Tiempo')

ax[0].set_ylabel('Temperatura')
plt.ylim(-20, 40)
plt.show()
# Close the netCDF4 file

